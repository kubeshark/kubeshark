package http

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

var http10protocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol -- HTTP/1.0",
	Abbreviation:    "HTTP",
	Macro:           "http",
	Version:         "1.0",
	BackgroundColor: "#205cf5",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://datatracker.ietf.org/doc/html/rfc1945",
	Ports:           []string{"80", "443", "8080"},
	Priority:        0,
}

var http11protocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol -- HTTP/1.1",
	Abbreviation:    "HTTP",
	Macro:           "http",
	Version:         "1.1",
	BackgroundColor: "#205cf5",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://datatracker.ietf.org/doc/html/rfc2616",
	Ports:           []string{"80", "443", "8080"},
	Priority:        0,
}

var http2Protocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol Version 2 (HTTP/2)",
	Abbreviation:    "HTTP/2",
	Macro:           "http2",
	Version:         "2.0",
	BackgroundColor: "#244c5a",
	ForegroundColor: "#ffffff",
	FontSize:        11,
	ReferenceLink:   "https://datatracker.ietf.org/doc/html/rfc7540",
	Ports:           []string{"80", "443", "8080"},
	Priority:        0,
}

var grpcProtocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol Version 2 (HTTP/2) [ gRPC over HTTP/2 ]",
	Abbreviation:    "gRPC",
	Macro:           "grpc",
	Version:         "2.0",
	BackgroundColor: "#244c5a",
	ForegroundColor: "#ffffff",
	FontSize:        11,
	ReferenceLink:   "https://grpc.github.io/grpc/core/md_doc__p_r_o_t_o_c_o_l-_h_t_t_p2.html",
	Ports:           []string{"80", "443", "8080", "50051"},
	Priority:        0,
}

var graphQL1Protocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol -- HTTP/1.1 [ GraphQL over HTTP/1.1 ]",
	Abbreviation:    "GQL",
	Macro:           "gql",
	Version:         "1.1",
	BackgroundColor: "#e10098",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://graphql.org/learn/serving-over-http/",
	Ports:           []string{"80", "443", "8080"},
	Priority:        0,
}

var graphQL2Protocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol Version 2 (HTTP/2) [ GraphQL over HTTP/2 ]",
	Abbreviation:    "GQL",
	Macro:           "gql",
	Version:         "2.0",
	BackgroundColor: "#e10098",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://graphql.org/learn/serving-over-http/",
	Ports:           []string{"80", "443", "8080", "50051"},
	Priority:        0,
}

const (
	TypeHttpRequest = iota
	TypeHttpResponse
)

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &http11protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s", http11protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, reader api.TcpReader, options *api.TrafficFilteringOptions) error {
	reqResMatcher := reader.GetReqResMatcher().(*requestResponseMatcher)

	var err error
	isHTTP2, _ := checkIsHTTP2Connection(b, reader.GetIsClient())

	var http2Assembler *Http2Assembler
	if isHTTP2 {
		err = prepareHTTP2Connection(b, reader.GetIsClient())
		if err != nil {
			return err
		}
		http2Assembler = createHTTP2Assembler(b)
	}

	switchingProtocolsHTTP2 := false
	for {
		if switchingProtocolsHTTP2 {
			switchingProtocolsHTTP2 = false
			isHTTP2, err = checkIsHTTP2Connection(b, reader.GetIsClient())
			if err != nil {
				break
			}
			err = prepareHTTP2Connection(b, reader.GetIsClient())
			if err != nil {
				break
			}
			http2Assembler = createHTTP2Assembler(b)
		}

		if isHTTP2 {
			err = handleHTTP2Stream(http2Assembler, reader.GetReadProgress(), reader.GetParent().GetOrigin(), reader.GetTcpID(), reader.GetCaptureTime(), reader.GetEmitter(), options, reqResMatcher)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				continue
			}
			reader.GetParent().SetProtocol(&http11protocol)
		} else if reader.GetIsClient() {
			var req *http.Request
			switchingProtocolsHTTP2, req, err = handleHTTP1ClientStream(b, reader.GetReadProgress(), reader.GetParent().GetOrigin(), reader.GetTcpID(), reader.GetCounterPair(), reader.GetCaptureTime(), reader.GetEmitter(), options, reqResMatcher)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				continue
			}
			reader.GetParent().SetProtocol(&http11protocol)

			// In case of an HTTP2 upgrade, duplicate the HTTP1 request into HTTP2 with stream ID 1
			if switchingProtocolsHTTP2 {
				ident := fmt.Sprintf(
					"%s_%s_%s_%s_1_%s",
					reader.GetTcpID().SrcIP,
					reader.GetTcpID().DstIP,
					reader.GetTcpID().SrcPort,
					reader.GetTcpID().DstPort,
					"HTTP2",
				)
				item := reqResMatcher.registerRequest(ident, req, reader.GetCaptureTime(), reader.GetReadProgress().Current(), req.ProtoMinor)
				if item != nil {
					item.ConnectionInfo = &api.ConnectionInfo{
						ClientIP:   reader.GetTcpID().SrcIP,
						ClientPort: reader.GetTcpID().SrcPort,
						ServerIP:   reader.GetTcpID().DstIP,
						ServerPort: reader.GetTcpID().DstPort,
						IsOutgoing: true,
					}
					item.Capture = reader.GetParent().GetOrigin()
					filterAndEmit(item, reader.GetEmitter(), options)
				}
			}
		} else {
			switchingProtocolsHTTP2, err = handleHTTP1ServerStream(b, reader.GetReadProgress(), reader.GetParent().GetOrigin(), reader.GetTcpID(), reader.GetCounterPair(), reader.GetCaptureTime(), reader.GetEmitter(), options, reqResMatcher)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				continue
			}
			reader.GetParent().SetProtocol(&http11protocol)
		}
	}

	return nil
}

func (d dissecting) Analyze(item *api.OutputChannelItem, resolvedSource string, resolvedDestination string, namespace string) *api.Entry {
	var host, authority, path string

	request := item.Pair.Request.Payload.(map[string]interface{})
	response := item.Pair.Response.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	resDetails := response["details"].(map[string]interface{})

	isRequestUpgradedH2C := false

	for _, header := range reqDetails["headers"].([]interface{}) {
		h := header.(map[string]interface{})
		if h["name"] == "Host" {
			host = h["value"].(string)
		}
		if h["name"] == ":authority" {
			authority = h["value"].(string)
		}
		if h["name"] == ":path" {
			path = h["value"].(string)
		}

		if h["name"] == "Upgrade" {
			if h["value"].(string) == "h2c" {
				isRequestUpgradedH2C = true
			}
		}
	}

	if isGraphQL(reqDetails) {
		if item.Protocol.Version == "2.0" {
			item.Protocol = graphQL2Protocol
		} else {
			item.Protocol = graphQL1Protocol
		}
	}

	if resDetails["bodySize"].(float64) < 0 {
		resDetails["bodySize"] = 0
	}

	if item.Protocol.Version == "2.0" && !isRequestUpgradedH2C {
		if resolvedDestination == "" {
			resolvedDestination = authority
		}
		if resolvedDestination == "" {
			resolvedDestination = host
		}
	} else {
		u, err := url.Parse(reqDetails["url"].(string))
		if err != nil {
			path = reqDetails["url"].(string)
		} else {
			path = u.Path
		}
	}

	request["url"] = reqDetails["url"].(string)
	reqDetails["targetUri"] = reqDetails["url"]
	reqDetails["path"] = path
	reqDetails["pathSegments"] = strings.Split(path, "/")[1:]

	// Rearrange the maps for the querying
	reqDetails["_headers"] = reqDetails["headers"]
	reqDetails["headers"] = mapSliceRebuildAsMap(reqDetails["_headers"].([]interface{}))
	resDetails["_headers"] = resDetails["headers"]
	resDetails["headers"] = mapSliceRebuildAsMap(resDetails["_headers"].([]interface{}))

	reqDetails["_cookies"] = reqDetails["cookies"]
	reqDetails["cookies"] = mapSliceRebuildAsMap(reqDetails["_cookies"].([]interface{}))
	resDetails["_cookies"] = resDetails["cookies"]
	resDetails["cookies"] = mapSliceRebuildAsMap(resDetails["_cookies"].([]interface{}))

	reqDetails["_queryString"] = reqDetails["queryString"]
	reqDetails["_queryStringMerged"] = mapSliceMergeRepeatedKeys(reqDetails["_queryString"].([]interface{}))
	reqDetails["queryString"] = mapSliceRebuildAsMap(reqDetails["_queryStringMerged"].([]interface{}))

	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	if elapsedTime < 0 {
		elapsedTime = 0
	}
	httpPair, _ := json.Marshal(item.Pair)
	return &api.Entry{
		Protocol: item.Protocol,
		Capture:  item.Capture,
		Source: &api.TCP{
			Name: resolvedSource,
			IP:   item.ConnectionInfo.ClientIP,
			Port: item.ConnectionInfo.ClientPort,
		},
		Destination: &api.TCP{
			Name: resolvedDestination,
			IP:   item.ConnectionInfo.ServerIP,
			Port: item.ConnectionInfo.ServerPort,
		},
		Namespace:    namespace,
		Outgoing:     item.ConnectionInfo.IsOutgoing,
		Request:      reqDetails,
		Response:     resDetails,
		RequestSize:  item.Pair.Request.CaptureSize,
		ResponseSize: item.Pair.Response.CaptureSize,
		Timestamp:    item.Timestamp,
		StartTime:    item.Pair.Request.CaptureTime,
		ElapsedTime:  elapsedTime,
		HTTPPair:     string(httpPair),
	}
}

func (d dissecting) Summarize(entry *api.Entry) *api.BaseEntry {
	summary := entry.Request["path"].(string)
	summaryQuery := fmt.Sprintf(`request.path == "%s"`, summary)
	method := entry.Request["method"].(string)
	methodQuery := fmt.Sprintf(`request.method == "%s"`, method)
	status := int(entry.Response["status"].(float64))
	statusQuery := fmt.Sprintf(`response.status == %d`, status)

	return &api.BaseEntry{
		Id:             entry.Id,
		Protocol:       entry.Protocol,
		Capture:        entry.Capture,
		Summary:        summary,
		SummaryQuery:   summaryQuery,
		Status:         status,
		StatusQuery:    statusQuery,
		Method:         method,
		MethodQuery:    methodQuery,
		Timestamp:      entry.Timestamp,
		Source:         entry.Source,
		Destination:    entry.Destination,
		IsOutgoing:     entry.Outgoing,
		Latency:        entry.ElapsedTime,
		Rules:          entry.Rules,
		ContractStatus: entry.ContractStatus,
	}
}

func representRequest(request map[string]interface{}) (repRequest []interface{}) {
	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Method",
			Value:    request["method"].(string),
			Selector: `request.method`,
		},
		{
			Name:     "Target URI",
			Value:    request["targetUri"].(string),
			Selector: `request.targetUri`,
		},
		{
			Name:     "Path",
			Value:    request["path"].(string),
			Selector: `request.path`,
		},
		{
			Name:     "Body Size (bytes)",
			Value:    int64(request["bodySize"].(float64)),
			Selector: `request.bodySize`,
		},
	})
	repRequest = append(repRequest, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	pathSegments := request["pathSegments"].([]interface{})
	if len(pathSegments) > 1 {
		repRequest = append(repRequest, api.SectionData{
			Type:  api.TABLE,
			Title: "Path Segments",
			Data:  representSliceAsTable(pathSegments, `request.pathSegments`),
		})
	}

	repRequest = append(repRequest, api.SectionData{
		Type:  api.TABLE,
		Title: "Headers",
		Data:  representMapSliceAsTable(request["_headers"].([]interface{}), `request.headers`),
	})

	repRequest = append(repRequest, api.SectionData{
		Type:  api.TABLE,
		Title: "Cookies",
		Data:  representMapSliceAsTable(request["_cookies"].([]interface{}), `request.cookies`),
	})

	repRequest = append(repRequest, api.SectionData{
		Type:  api.TABLE,
		Title: "Query String",
		Data:  representMapSliceAsTable(request["_queryStringMerged"].([]interface{}), `request.queryString`),
	})

	postData, _ := request["postData"].(map[string]interface{})
	mimeType := postData["mimeType"]
	if mimeType == nil {
		mimeType = ""
	}
	text := postData["text"]
	if text != nil {
		repRequest = append(repRequest, api.SectionData{
			Type:     api.BODY,
			Title:    "POST Data (text/plain)",
			MimeType: mimeType.(string),
			Data:     text.(string),
			Selector: `request.postData.text`,
		})
	}

	if postData["params"] != nil {
		params, _ := json.Marshal(postData["params"].([]interface{}))
		if len(params) > 0 {
			if mimeType == "multipart/form-data" {
				multipart, _ := json.Marshal([]map[string]string{
					{
						"name":  "Files",
						"value": string(params),
					},
				})
				repRequest = append(repRequest, api.SectionData{
					Type:  api.TABLE,
					Title: "POST Data (multipart/form-data)",
					Data:  string(multipart),
				})
			} else {
				repRequest = append(repRequest, api.SectionData{
					Type:  api.TABLE,
					Title: "POST Data (application/x-www-form-urlencoded)",
					Data:  representMapSliceAsTable(postData["params"].([]interface{}), `request.postData.params`),
				})
			}
		}
	}

	return
}

func representResponse(response map[string]interface{}) (repResponse []interface{}) {
	repResponse = make([]interface{}, 0)

	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Status",
			Value:    int64(response["status"].(float64)),
			Selector: `response.status`,
		},
		{
			Name:     "Status Text",
			Value:    response["statusText"].(string),
			Selector: `response.statusText`,
		},
		{
			Name:     "Body Size (bytes)",
			Value:    int64(response["bodySize"].(float64)),
			Selector: `response.bodySize`,
		},
	})
	repResponse = append(repResponse, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	repResponse = append(repResponse, api.SectionData{
		Type:  api.TABLE,
		Title: "Headers",
		Data:  representMapSliceAsTable(response["_headers"].([]interface{}), `response.headers`),
	})

	repResponse = append(repResponse, api.SectionData{
		Type:  api.TABLE,
		Title: "Cookies",
		Data:  representMapSliceAsTable(response["_cookies"].([]interface{}), `response.cookies`),
	})

	content, _ := response["content"].(map[string]interface{})
	mimeType := content["mimeType"]
	if mimeType == nil {
		mimeType = ""
	}
	encoding := content["encoding"]
	text := content["text"]
	if text != nil {
		repResponse = append(repResponse, api.SectionData{
			Type:     api.BODY,
			Title:    "Body",
			Encoding: encoding.(string),
			MimeType: mimeType.(string),
			Data:     text.(string),
			Selector: `response.content.text`,
		})
	}

	return
}

func (d dissecting) Represent(request map[string]interface{}, response map[string]interface{}) (object []byte, err error) {
	representation := make(map[string]interface{})
	repRequest := representRequest(request)
	repResponse := representResponse(response)
	representation["request"] = repRequest
	representation["response"] = repResponse
	object, err = json.Marshal(representation)
	return
}

func (d dissecting) Macros() map[string]string {
	return map[string]string{
		`http`:  fmt.Sprintf(`proto.name == "%s" and proto.version.startsWith("%c")`, http11protocol.Name, http11protocol.Version[0]),
		`http2`: fmt.Sprintf(`proto.name == "%s" and proto.version == "%s"`, http11protocol.Name, http2Protocol.Version),
		`grpc`:  fmt.Sprintf(`proto.name == "%s" and proto.version == "%s" and proto.macro == "%s"`, http11protocol.Name, grpcProtocol.Version, grpcProtocol.Macro),
		`gql`:   fmt.Sprintf(`proto.name == "%s" and proto.macro == "%s"`, graphQL1Protocol.Name, graphQL1Protocol.Macro),
	}
}

func (d dissecting) NewResponseRequestMatcher() api.RequestResponseMatcher {
	return createResponseRequestMatcher()
}

var Dissector dissecting

func NewDissector() api.Dissector {
	return Dissector
}
