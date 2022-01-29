package oas

import (
	"encoding/json"
	"errors"
	"github.com/chanced/openapi"
	"github.com/google/uuid"
	"github.com/nav-inc/datetime"
	"github.com/up9inc/mizu/shared/logger"
	"mime"
	"mizuserver/pkg/har"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const CountersTotal = "x-counters-total"
const CountersPerSource = "x-counters-per-source"

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

func (g *SpecGen) feedEntry(entryWithSource EntryWithSource) (string, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	opId, err := g.handlePathObj(&entryWithSource)
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

	counterTotal := Counter{}
	counterMapTotal := CounterMap{}

	for _, pathop := range g.tree.listOps() {
		opObj := pathop.op
		if opObj.Summary == "" {
			opObj.Summary = pathop.path
		}

		if _, ok := opObj.Extensions.Extension(CountersTotal); ok {
			counter := new(Counter)
			err := opObj.Extensions.DecodeExtension(CountersTotal, counter)
			if err != nil {
				return nil, err
			}
			counterTotal.addOther(counter)
		}

		if _, ok := opObj.Extensions.Extension(CountersPerSource); ok {
			counterMap := new(CounterMap)
			err := opObj.Extensions.DecodeExtension(CountersPerSource, counterMap)
			if err != nil {
				return nil, err
			}
			counterMapTotal.addOther(counterMap)
		}
	}

	if g.oas.Extensions == nil {
		g.oas.Extensions = openapi.Extensions{}
	}

	err := g.oas.Extensions.SetExtension(CountersTotal, counterTotal)
	if err != nil {
		return nil, err
	}

	err = g.oas.Extensions.SetExtension(CountersPerSource, counterMapTotal)
	if err != nil {
		return nil, err
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

		//groups[common] = group
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

func (g *SpecGen) handlePathObj(entryWithSource *EntryWithSource) (string, error) {
	entry := entryWithSource.Entry
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
	opObj, err := handleOpObj(entryWithSource, node.pathObj)

	if opObj != nil {
		return opObj.OperationID, err
	}

	return "", err
}

func handleOpObj(entryWithSource *EntryWithSource, pathObj *openapi.PathObj) (*openapi.Operation, error) {
	entry := entryWithSource.Entry
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

	err = handleCounters(opObj, isSuccess, entryWithSource)
	if err != nil {
		return nil, err
	}

	return opObj, nil
}

func handleCounters(opObj *openapi.Operation, success bool, entryWithSource *EntryWithSource) error {
	counter := Counter{}
	counterMap := CounterMap{}
	if opObj.Extensions == nil {
		opObj.Extensions = openapi.Extensions{}
	} else {
		if _, ok := opObj.Extensions.Extension(CountersTotal); ok {
			err := opObj.Extensions.DecodeExtension(CountersTotal, &counter)
			if err != nil {
				return err
			}
		}

		if _, ok := opObj.Extensions.Extension(CountersPerSource); ok {
			err := opObj.Extensions.DecodeExtension(CountersPerSource, &counterMap)
			if err != nil {
				return err
			}
		}
	}

	var counterPerSource *Counter
	if existing, ok := counterMap[entryWithSource.Source]; ok {
		counterPerSource = existing
	} else {
		counterPerSource = new(Counter)
		counterMap[entryWithSource.Source] = counterPerSource
	}

	started, err := datetime.Parse(entryWithSource.Entry.StartedDateTime, time.UTC)
	if err != nil {
		return err
	}

	ts := float64(started.UnixMilli()) / 1000
	rt := float64(entryWithSource.Entry.Time) / 1000

	counter.addEntry(ts, rt, success)
	counterPerSource.addEntry(ts, rt, success)

	err = opObj.Extensions.SetExtension(CountersTotal, counter)
	if err != nil {
		return err
	}

	err = opObj.Extensions.SetExtension(CountersPerSource, counterMap)
	if err != nil {
		return err
	}

	return nil
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
			reqCtype := getReqCtype(req)
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

		content.Example = exampleMsg
	}

	return respContent[ctype], nil
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

func getReqCtype(req *har.Request) string {
	var ctype string
	ctype = req.PostData.MimeType
	for _, hdr := range req.Headers {
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
