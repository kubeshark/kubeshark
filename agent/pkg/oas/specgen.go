package oas

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/chanced/openapi"
	"github.com/google/martian/har"
	"github.com/google/uuid"
	"github.com/up9inc/mizu/shared/logger"
	"mime"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

var IgnoredExtensions = []string{"gif", "svg", "css", "png", "ico", "js", "woff2", "woff", "jpg", "jpeg", "swf", "ttf", "map", "webp", "otf"}
var IgnoredCtypePrefixes = []string{"image/", "font/", "video/", "text/javascript"}
var IgnoredCtypes = []string{"application/javascript", "application/x-javascript", "text/css", "application/font-woff2", "application/font-woff", "application/x-font-woff"}
var IgnoredHeaders = []string{
	"a-im", "accept",
	"authorization", "cache-control", "connection", "content-encoding", "content-length", "content-type", "cookie",
	"date", "dnt", "expect", "forwarded", "from", "front-end-https", "host", "http2-settings",
	"max-forwards", "origin", "pragma", "proxy-authorization", "proxy-connection", "range", "referer",
	"save-data", "te", "trailer", "transfer-encoding", "upgrade", "upgrade-insecure-requests",
	"server", "user-agent", "via", "warning", "strict-transport-security",
	"x-att-deviceid", "x-correlation-id", "correlation-id", "x-client-data",
	"x-http-method-override", "x-real-ip", "x-request-id", "x-request-start", "x-requested-with", "x-uidh",
	"x-same-domain", "x-content-type-options", "x-frame-options", "x-xss-protection",
	"x-wap-profile", "x-scheme",
	"newrelic", "x-cloud-trace-context", "sentry-trace",
	"expires", "set-cookie", "p3p", "location", "content-security-policy", "content-security-policy-report-only",
	"last-modified", "content-language",
	"keep-alive", "etag", "alt-svc", "x-csrf-token", "x-ua-compatible", "vary", "x-powered-by",
	"age", "allow", "www-authenticate",
}

var IgnoredHeaderPrefixes = []string{
	":", "accept-", "access-control-", "if-", "sec-", "grpc-",
	"x-forwarded-", "x-original-",
	"x-up9-", "x-envoy-", "x-hasura-", "x-b3-", "x-datadog-", "x-envoy-", "x-amz-", "x-newrelic-", "x-prometheus-",
	"x-akamai-", "x-spotim-", "x-amzn-", "x-ratelimit-",
}

func isCtypeIgnored(ctype string) bool {
	for _, prefix := range IgnoredCtypePrefixes {
		if strings.HasPrefix(ctype, prefix) {
			return true
		}
	}

	for _, toIgnore := range IgnoredCtypes {
		if ctype == toIgnore {
			return true
		}
	}
	return false
}

func isExtIgnored(path string) bool {
	for _, extIgn := range IgnoredExtensions {
		if strings.HasSuffix(path, "."+extIgn) {
			return true
		}
	}
	return false
}

type ReqResp struct { // hello, generics in Go
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
	info.Version = "0.0"
	spec.Info = &info
	spec.Paths = &openapi.Paths{Items: map[openapi.PathValue]*openapi.PathObj{}}
	gen := SpecGen{oas: spec, tree: new(Node)}
	return &gen
}

func (g *SpecGen) startFromSpec(oas *openapi.OpenAPI) {
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

	return opId, err
}

func (g *SpecGen) GetSpec() (*openapi.OpenAPI, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.tree.compact()

	// put paths back from tree into OAS
	g.oas.Paths = g.tree.listPaths()

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

func (g *SpecGen) handlePathObj(entry *har.Entry) (string, error) {
	urlParsed, err := url.Parse(entry.Request.URL)
	if err != nil {
		return "", err
	}

	if isExtIgnored(urlParsed.Path) {
		logger.Log.Debugf("Dropped traffic entry due to ignored extension: %s", urlParsed.Path)
	}

	ctype := getRespCtype(entry.Response)
	if isCtypeIgnored(ctype) {
		logger.Log.Debugf("Dropped traffic entry due to ignored response ctype: %s", ctype)
	}

	if entry.Response.Status < 100 {
		logger.Log.Debugf("Dropped traffic entry due to status<100: %s", entry.StartedDateTime)
		return "", nil
	}

	if entry.Response.Status == 301 || entry.Response.Status == 308 {
		logger.Log.Debugf("Dropped traffic entry due to permanent redirect status: %s", entry.StartedDateTime)
		return "", nil
	}

	split := strings.Split(urlParsed.Path, "/")
	node := g.tree.getOrSet(split, new(openapi.PathObj))
	opObj, err := handleOpObj(entry, node.ops)

	return opObj.OperationID, err
}

func (g *SpecGen) getPathObj(entryURL string) (*openapi.PathObj, error) {
	urlParsed, err := url.Parse(entryURL)
	if err != nil {
		return nil, err
	}

	urlPath := openapi.PathValue(urlParsed.Path)

	var pathObj *openapi.PathObj
	for k, v := range g.oas.Paths.Items {
		if k == urlPath {
			pathObj = v
		}
	}

	if pathObj == nil {
		g.oas.Paths.Items[urlPath] = &openapi.PathObj{}
		pathObj = g.oas.Paths.Items[urlPath]
	}

	return pathObj, nil
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

	err = handleRequest(entry.Request, opObj, isSuccess)
	if err != nil {
		return nil, err
	}

	err = handleResponse(entry.Response, opObj, isSuccess)
	if err != nil {
		return nil, err
	}

	return opObj, nil
}

func handleRequest(req *har.Request, opObj *openapi.Operation, isSuccess bool) error {
	if req.PostData != nil && req.PostData.Text != "" && isSuccess {
		reqBody, err := getRequestBody(req, opObj, isSuccess)
		if err != nil {
			return err
		}

		if reqBody != nil {
			reqCtype := getReqCtype(req)
			reqMedia, err := fillContent(ReqResp{Req: req}, reqBody.Content, reqCtype, err)
			if err != nil {
				return err
			}

			_ = reqMedia
		}
	}
	return nil
}

func handleResponse(resp *har.Response, opObj *openapi.Operation, isSuccess bool) error {
	respObj, err := getResponseObj(resp, opObj, isSuccess)
	if err != nil {
		return err
	}

	respCtype := getRespCtype(resp)
	respContent := respObj.Content
	respMedia, err := fillContent(ReqResp{Resp: resp}, respContent, respCtype, err)
	if err != nil {
		return err
	}
	_ = respMedia
	return nil
}

func fillContent(reqResp ReqResp, respContent openapi.Content, ctype string, err error) (*openapi.MediaType, error) {
	content, found := respContent[ctype]
	if !found {
		respContent[ctype] = &openapi.MediaType{}
		content = respContent[ctype]
	}

	var text string
	if reqResp.Req != nil {
		text = reqResp.Req.PostData.Text
	} else {
		text = decRespText(reqResp.Resp.Content)
	}

	exampleMsg, err := json.Marshal(text)
	if err != nil {
		return nil, err
	}
	content.Example = exampleMsg
	content.Example = nil
	return respContent[ctype], nil
}

func decRespText(content *har.Content) (res string) {
	res = string(content.Text)
	if content.Encoding == "base64" {
		data, err := base64.StdEncoding.DecodeString(res)
		if err != nil {
			logger.Log.Warningf("error decoding response text as base64: %s", err)
		} else {
			res = string(data)
		}
	}
	return
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
		opObj.Responses[statusStr] = &openapi.ResponseObj{Content: map[string]*openapi.MediaType{}}
		response = opObj.Responses[statusStr]
	}

	var resResponse *openapi.ResponseObj
	switch response.ResponseKind() {
	case openapi.ResponseKindRef:
		return nil, errors.New("response reference is not supported at the moment")
	case openapi.ResponseKindObj:
		resResponse = response.(*openapi.ResponseObj)
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

	var reqBody *openapi.RequestBodyObj

	switch opObj.RequestBody.RequestBodyKind() {
	case openapi.RequestBodyKindRef:
		return nil, errors.New("request body reference is not supported at the moment")
	case openapi.RequestBodyKindObj:
		reqBody = opObj.RequestBody.(*openapi.RequestBodyObj)
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
		return nil, false, errors.New("Unsupported HTTP method: " + method)
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
