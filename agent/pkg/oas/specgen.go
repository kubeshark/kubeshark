package oas

import (
	"encoding/json"
	"errors"
	"github.com/chanced/openapi"
	"github.com/google/uuid"
	"github.com/up9inc/mizu/shared/logger"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"mizuserver/pkg/har"
	"net/textproto"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type reqResp struct { // hello, generics in Go
	Req  *har.Request
	Resp *har.Response
}

type SpecGen struct {
	oas  *openapi.OpenAPI
	tree *Node
	lock sync.Mutex
}

func NewGen(server string) *SpecGen {
	spec := new(openapi.OpenAPI)
	spec.Version = "3.1.0"

	info := openapi.Info{Title: server}
	info.Version = "1.0"
	spec.Info = &info
	spec.Paths = &openapi.Paths{Items: map[openapi.PathValue]*openapi.PathObj{}}

	spec.Servers = make([]*openapi.Server, 0)
	spec.Servers = append(spec.Servers, &openapi.Server{URL: server})

	gen := SpecGen{oas: spec, tree: new(Node)}
	return &gen
}

func (g *SpecGen) StartFromSpec(oas *openapi.OpenAPI) {
	g.oas = oas
	for pathStr, pathObj := range oas.Paths.Items {
		pathSplit := strings.Split(string(pathStr), "/")
		g.tree.getOrSet(pathSplit, pathObj)
	}
}

func (g *SpecGen) feedEntry(entry har.Entry) (string, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	opId, err := g.handlePathObj(&entry)
	if err != nil {
		return "", err
	}

	// NOTE: opId can be empty for some failed entries
	return opId, err
}

func (g *SpecGen) GetSpec() (*openapi.OpenAPI, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.tree.compact()

	for _, pathop := range g.tree.listOps() {
		if pathop.op.Summary == "" {
			pathop.op.Summary = pathop.path
		}
	}

	// put paths back from tree into OAS
	g.oas.Paths = g.tree.listPaths()

	suggestTags(g.oas)

	// to make a deep copy, no better idea than marshal+unmarshal
	specText, err := json.MarshalIndent(g.oas, "", "\t")
	if err != nil {
		return nil, err
	}

	spec := new(openapi.OpenAPI)
	err = json.Unmarshal(specText, spec)
	if err != nil {
		return nil, err
	}

	return spec, err
}

func suggestTags(oas *openapi.OpenAPI) {
	paths := getPathsKeys(oas.Paths.Items)
	sort.Strings(paths) // make it stable in case of multiple candidates
	for len(paths) > 0 {
		group := make([]string, 0)
		group = append(group, paths[0])
		paths = paths[1:]

		pathsClone := append(paths[:0:0], paths...)
		for _, path := range pathsClone {
			if getSimilarPrefix([]string{group[0], path}) != "" {
				group = append(group, path)
				paths = deleteFromSlice(paths, path)
			}
		}

		common := getSimilarPrefix(group)

		if len(group) > 1 {
			for _, path := range group {
				pathObj := oas.Paths.Items[openapi.PathValue(path)]
				for _, op := range getOps(pathObj) {
					if op.Tags == nil {
						op.Tags = make([]string, 0)
					}
					// only add tags if not present
					if len(op.Tags) == 0 {
						op.Tags = append(op.Tags, common)
					}
				}
			}
		}
	}
}

func getSimilarPrefix(strs []string) string {
	chunked := make([][]string, 0)
	for _, item := range strs {
		chunked = append(chunked, strings.Split(item, "/"))
	}

	cmn := longestCommonXfix(chunked, true)
	res := make([]string, 0)
	for _, chunk := range cmn {
		if chunk != "api" && !IsVersionString(chunk) && !strings.HasPrefix(chunk, "{") {
			res = append(res, chunk)
		}
	}
	return strings.Join(res[1:], ".")
}

func deleteFromSlice(s []string, val string) []string {
	temp := s[:0]
	for _, x := range s {
		if x != val {
			temp = append(temp, x)
		}
	}
	return temp
}

func getPathsKeys(mymap map[openapi.PathValue]*openapi.PathObj) []string {
	keys := make([]string, len(mymap))

	i := 0
	for k := range mymap {
		keys[i] = string(k)
		i++
	}
	return keys
}

func (g *SpecGen) handlePathObj(entry *har.Entry) (string, error) {
	urlParsed, err := url.Parse(entry.Request.URL)
	if err != nil {
		return "", err
	}

	if isExtIgnored(urlParsed.Path) {
		logger.Log.Debugf("Dropped traffic entry due to ignored extension: %s", urlParsed.Path)
		return "", nil
	}

	if entry.Request.Method == "OPTIONS" {
		logger.Log.Debugf("Dropped traffic entry due to its method: %s", urlParsed.Path)
		return "", nil
	}

	ctype := getRespCtype(&entry.Response)
	if isCtypeIgnored(ctype) {
		logger.Log.Debugf("Dropped traffic entry due to ignored response ctype: %s", ctype)
		return "", nil
	}

	if entry.Response.Status < 100 {
		logger.Log.Debugf("Dropped traffic entry due to status<100: %s", entry.StartedDateTime)
		return "", nil
	}

	if entry.Response.Status == 301 || entry.Response.Status == 308 {
		logger.Log.Debugf("Dropped traffic entry due to permanent redirect status: %s", entry.StartedDateTime)
		return "", nil
	}

	if entry.Response.Status == 502 || entry.Response.Status == 503 || entry.Response.Status == 504 {
		logger.Log.Debugf("Dropped traffic entry due to temporary server error: %s", entry.StartedDateTime)
		return "", nil
	}

	var split []string
	if urlParsed.RawPath != "" {
		split = strings.Split(urlParsed.RawPath, "/")
	} else {
		split = strings.Split(urlParsed.Path, "/")
	}
	node := g.tree.getOrSet(split, new(openapi.PathObj))
	opObj, err := handleOpObj(entry, node.pathObj)

	if opObj != nil {
		return opObj.OperationID, err
	}

	return "", err
}

func handleOpObj(entry *har.Entry, pathObj *openapi.PathObj) (*openapi.Operation, error) {
	isSuccess := 100 <= entry.Response.Status && entry.Response.Status < 400
	opObj, wasMissing, err := getOpObj(pathObj, entry.Request.Method, isSuccess)
	if err != nil {
		return nil, err
	}

	if !isSuccess && wasMissing {
		logger.Log.Debugf("Dropped traffic entry due to failed status and no known endpoint at: %s", entry.StartedDateTime)
		return nil, nil
	}

	err = handleRequest(&entry.Request, opObj, isSuccess)
	if err != nil {
		return nil, err
	}

	err = handleResponse(&entry.Response, opObj, isSuccess)
	if err != nil {
		return nil, err
	}

	return opObj, nil
}

func handleRequest(req *har.Request, opObj *openapi.Operation, isSuccess bool) error {
	// TODO: we don't handle the situation when header/qstr param can be defined on pathObj level. Also the path param defined on opObj

	qstrGW := nvParams{
		In:             openapi.InQuery,
		Pairs:          req.QueryString,
		IsIgnored:      func(name string) bool { return false },
		GeneralizeName: func(name string) string { return name },
	}
	handleNameVals(qstrGW, &opObj.Parameters)

	hdrGW := nvParams{
		In:             openapi.InHeader,
		Pairs:          req.Headers,
		IsIgnored:      isHeaderIgnored,
		GeneralizeName: strings.ToLower,
	}
	handleNameVals(hdrGW, &opObj.Parameters)

	if req.PostData.Text != "" && isSuccess {
		reqBody, err := getRequestBody(req, opObj, isSuccess)
		if err != nil {
			return err
		}

		if reqBody != nil {
			reqCtype, _ := getReqCtype(req)
			reqMedia, err := fillContent(reqResp{Req: req}, reqBody.Content, reqCtype, err)
			if err != nil {
				return err
			}

			_ = reqMedia
		}
	}
	return nil
}

func handleResponse(resp *har.Response, opObj *openapi.Operation, isSuccess bool) error {
	// TODO: we don't support "default" response
	respObj, err := getResponseObj(resp, opObj, isSuccess)
	if err != nil {
		return err
	}

	handleRespHeaders(resp.Headers, respObj)

	respCtype := getRespCtype(resp)
	respContent := respObj.Content
	respMedia, err := fillContent(reqResp{Resp: resp}, respContent, respCtype, err)
	if err != nil {
		return err
	}
	_ = respMedia
	return nil
}

func handleRespHeaders(reqHeaders []har.Header, respObj *openapi.ResponseObj) {
	visited := map[string]*openapi.HeaderObj{}
	for _, pair := range reqHeaders {
		if isHeaderIgnored(pair.Name) {
			continue
		}

		nameGeneral := strings.ToLower(pair.Name)

		initHeaders(respObj)
		objHeaders := respObj.Headers
		param := findHeaderByName(&respObj.Headers, pair.Name)
		if param == nil {
			param = createHeader(openapi.TypeString)
			objHeaders[nameGeneral] = param
		}
		exmp := &param.Examples
		err := fillParamExample(&exmp, pair.Value)
		if err != nil {
			logger.Log.Warningf("Failed to add example to a parameter: %s", err)
		}
		visited[nameGeneral] = param
	}

	// maintain "required" flag
	if respObj.Headers != nil {
		for name, param := range respObj.Headers {
			paramObj, err := param.ResolveHeader(headerResolver)
			if err != nil {
				logger.Log.Warningf("Failed to resolve param: %s", err)
				continue
			}

			_, ok := visited[strings.ToLower(name)]
			if !ok {
				flag := false
				paramObj.Required = &flag
			}
		}
	}

	return
}

func fillContent(reqResp reqResp, respContent openapi.Content, ctype string, err error) (*openapi.MediaType, error) {
	content, found := respContent[ctype]
	if !found {
		respContent[ctype] = &openapi.MediaType{}
		content = respContent[ctype]
	}

	var text string
	var isBinary bool
	if reqResp.Req != nil {
		isBinary, _, text = reqResp.Req.PostData.B64Decoded()
	} else {
		isBinary, _, text = reqResp.Resp.Content.B64Decoded()
	}

	if !isBinary && text != "" {
		var exampleMsg []byte
		// try treating it as json
		any, isJSON := anyJSON(text)
		if isJSON {
			// re-marshal with forced indent
			exampleMsg, err = json.MarshalIndent(any, "", "\t")
			if err != nil {
				panic("Failed to re-marshal value, super-strange")
			}
		} else {
			exampleMsg, err = json.Marshal(text)
			if err != nil {
				return nil, err
			}
		}

		if ctype == "application/x-www-form-urlencoded" && reqResp.Req != nil {
			handleFormDataUrlencoded(text, content)
		} else if strings.HasPrefix(ctype, "multipart/form-data") && reqResp.Req != nil {
			_, params := getReqCtype(reqResp.Req)
			handleFormDataMultipart(text, content, params)
		}

		if content.Example == nil && len(exampleMsg) > len(content.Example) {
			content.Example = exampleMsg
		}
	}

	return respContent[ctype], nil
}

func handleFormDataUrlencoded(text string, content *openapi.MediaType) {
	formData, err := url.ParseQuery(text)
	if err != nil {
		logger.Log.Warningf("Could not decode urlencoded: %s", err)
		return
	}

	parts := make([]PartWithBody, 0)
	for name, vals := range formData {
		for _, val := range vals {
			part := new(multipart.Part)
			part.Header = textproto.MIMEHeader{}
			if name != part.FormName() {
				panic("")
			}
			parts = append(parts, PartWithBody{part: part, body: []byte(val)})
		}
	}
	handleFormData(text, content, parts)
}

func handleFormData(text string, content *openapi.MediaType, parts []PartWithBody) {
	if content.Schema == nil {
		content.Schema = new(openapi.SchemaObj)
		content.Schema.Type = openapi.Types{openapi.TypeObject}
		content.Schema.Properties = openapi.Schemas{}
	}

	props := &content.Schema.Properties
	for _, pwb := range parts {
		name := pwb.part.FormName()
		existing, found := (*props)[name]
		if !found {
			existing = new(openapi.SchemaObj)
			existing.Type = openapi.Types{openapi.TypeString}
			(*props)[name] = existing
		}

		examples := make([]string, 0)
		if existing.Examples != nil {
			err := json.Unmarshal(existing.Examples, &examples)
			if err != nil {
				continue
			}
		}

		if len(examples) < 5 {
		byVals:
			for _, val := range vals {
				for _, eVal := range examples {
					if eVal == val {
						continue byVals
					}
				}
				examples = append(examples, val)
			}
		}

		raw, err := json.Marshal(examples)
		if err != nil {
			continue
		}
		existing.Examples = raw
	}
}

type PartWithBody struct {
	part *multipart.Part
	body []byte
}

func handleFormDataMultipart(text string, content *openapi.MediaType, ctypeParams map[string]string) {
	boundary, ok := ctypeParams["boundary"]
	if !ok {
		logger.Log.Errorf("Multipart header has no boundary")
		return
	}
	mpr := multipart.NewReader(strings.NewReader(text), boundary)

	parts := make([]PartWithBody, 0)
	for {
		part, err := mpr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Log.Errorf("Cannot parse multipart body: %v", err)
			break
		}
		defer part.Close()

		body, err := ioutil.ReadAll(part)
		if err != nil {
			logger.Log.Errorf("Error reading multipart Part %s: %v", part.Header, err)
		}

		parts = append(parts, PartWithBody{part: part, body: body})
	}

	handleFormData(text, content, parts)
}

func getRespCtype(resp *har.Response) string {
	var ctype string
	ctype = resp.Content.MimeType
	for _, hdr := range resp.Headers {
		if strings.ToLower(hdr.Name) == "content-type" {
			ctype = hdr.Value
		}
	}

	mediaType, _, err := mime.ParseMediaType(ctype)
	if err != nil {
		return ""
	}
	return mediaType
}

func getReqCtype(req *har.Request) (ctype string, params map[string]string) {
	ctype = req.PostData.MimeType
	for _, hdr := range req.Headers {
		if strings.ToLower(hdr.Name) == "content-type" {
			ctype = hdr.Value
		}
	}

	mediaType, params, err := mime.ParseMediaType(ctype)
	if err != nil {
		logger.Log.Errorf("Cannot parse Content-Type header %q: %v", ctype, err)
		return "", map[string]string{}
	}
	return mediaType, params
}

func getResponseObj(resp *har.Response, opObj *openapi.Operation, isSuccess bool) (*openapi.ResponseObj, error) {
	statusStr := strconv.Itoa(resp.Status)

	var response openapi.Response
	response, found := opObj.Responses[statusStr]
	if !found {
		if opObj.Responses == nil {
			opObj.Responses = map[string]openapi.Response{}
		}

		opObj.Responses[statusStr] = &openapi.ResponseObj{Content: map[string]*openapi.MediaType{}}
		response = opObj.Responses[statusStr]
	}

	resResponse, err := response.ResolveResponse(responseResolver)
	if err != nil {
		return nil, err
	}

	if isSuccess {
		resResponse.Description = "Successful call with status " + statusStr
	} else {
		resResponse.Description = "Failed call with status " + statusStr
	}
	return resResponse, nil
}

func getRequestBody(req *har.Request, opObj *openapi.Operation, isSuccess bool) (*openapi.RequestBodyObj, error) {
	if opObj.RequestBody == nil {
		opObj.RequestBody = &openapi.RequestBodyObj{Description: "Generic request body", Required: true, Content: map[string]*openapi.MediaType{}}
	}

	reqBody, err := opObj.RequestBody.ResolveRequestBody(reqBodyResolver)
	if err != nil {
		return nil, err
	}

	// TODO: maintain required flag for it, but only consider successful responses
	//reqBody.Content[]

	return reqBody, nil
}

func getOpObj(pathObj *openapi.PathObj, method string, createIfNone bool) (*openapi.Operation, bool, error) {
	method = strings.ToLower(method)
	var op **openapi.Operation

	switch method {
	case "get":
		op = &pathObj.Get
	case "put":
		op = &pathObj.Put
	case "post":
		op = &pathObj.Post
	case "delete":
		op = &pathObj.Delete
	case "options":
		op = &pathObj.Options
	case "head":
		op = &pathObj.Head
	case "patch":
		op = &pathObj.Patch
	case "trace":
		op = &pathObj.Trace
	default:
		return nil, false, errors.New("unsupported HTTP method: " + method)
	}

	isMissing := false
	if *op == nil {
		isMissing = true
		if createIfNone {
			*op = &openapi.Operation{Responses: map[string]openapi.Response{}}
			newUUID := uuid.New().String()
			(**op).OperationID = newUUID
		} else {
			return nil, isMissing, nil
		}
	}

	return *op, isMissing, nil
}
