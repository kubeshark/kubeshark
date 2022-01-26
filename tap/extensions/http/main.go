package http

import (
	"bufio"
	"encoding/json"
	"errors"
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
	ReferenceLink:   "https://grpc.github.io/grpc/core/md_doc_statuscodes.html",
	Ports:           []string{"80", "443", "8080", "50051"},
	Priority:        0,
}

const (
	TypeHttpRequest = iota
	TypeHttpResponse
)

func init() {
	log.Println("Initializing HTTP extension...")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &http11protocol
	extension.MatcherMap = reqResMatcher.openMessagesMap
}

func (d dissecting) Ping() {
	log.Printf("pong %s", http11protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, superIdentifier *api.SuperIdentifier, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	isHTTP2, err := checkIsHTTP2Connection(b, isClient)

	var http2Assembler *Http2Assembler
	if isHTTP2 {
		prepareHTTP2Connection(b, isClient)
		http2Assembler = createHTTP2Assembler(b)
	}

	dissected := false
	switchingProtocolsHTTP2 := false
	for {
		if switchingProtocolsHTTP2 {
			switchingProtocolsHTTP2 = false
			isHTTP2, err = checkIsHTTP2Connection(b, isClient)
			prepareHTTP2Connection(b, isClient)
			http2Assembler = createHTTP2Assembler(b)
		}

		if superIdentifier.Protocol != nil && superIdentifier.Protocol != &http11protocol {
			return errors.New("Identified by another protocol")
		}

		if isHTTP2 {
			err = handleHTTP2Stream(http2Assembler, tcpID, superTimer, emitter, options)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				continue
			}
			dissected = true
		} else if isClient {
			var req *http.Request
			switchingProtocolsHTTP2, req, err = handleHTTP1ClientStream(b, tcpID, counterPair, superTimer, emitter, options)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				continue
			}
			dissected = true

			// In case of an HTTP2 upgrade, duplicate the HTTP1 request into HTTP2 with stream ID 1
			if switchingProtocolsHTTP2 {
				ident := fmt.Sprintf(
					"%s->%s %s->%s 1 %s",
					tcpID.SrcIP,
					tcpID.DstIP,
					tcpID.SrcPort,
					tcpID.DstPort,
					"HTTP2",
				)
				item := reqResMatcher.registerRequest(ident, req, superTimer.CaptureTime, req.ProtoMinor)
				if item != nil {
					item.ConnectionInfo = &api.ConnectionInfo{
						ClientIP:   tcpID.SrcIP,
						ClientPort: tcpID.SrcPort,
						ServerIP:   tcpID.DstIP,
						ServerPort: tcpID.DstPort,
						IsOutgoing: true,
					}
					filterAndEmit(item, emitter, options)
				}
			}
		} else {
			switchingProtocolsHTTP2, err = handleHTTP1ServerStream(b, tcpID, counterPair, superTimer, emitter, options)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				continue
			}
			dissected = true
		}
	}

	if !dissected {
		return err
	}
	superIdentifier.Protocol = &http11protocol
	return nil
}

func (d dissecting) Analyze(item *api.OutputChannelItem, resolvedSource string, resolvedDestination string) *api.Entry {
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
	reqDetails["summary"] = path

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

	method := reqDetails["method"].(string)
	statusCode := int(resDetails["status"].(float64))
	if item.Protocol.Abbreviation == "gRPC" {
		resDetails["statusText"] = grpcStatusCodes[statusCode]
	}

	if item.Protocol.Version == "2.0" && !isRequestUpgradedH2C {
		reqDetails["url"] = path
		request["url"] = path
	}

	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	if elapsedTime < 0 {
		elapsedTime = 0
	}
	httpPair, _ := json.Marshal(item.Pair)
	return &api.Entry{
		Protocol: item.Protocol,
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
		Outgoing:    item.ConnectionInfo.IsOutgoing,
		Request:     reqDetails,
		Response:    resDetails,
		Method:      method,
		Status:      statusCode,
		Timestamp:   item.Timestamp,
		StartTime:   item.Pair.Request.CaptureTime,
		ElapsedTime: elapsedTime,
		Summary:     path,
		IsOutgoing:  item.ConnectionInfo.IsOutgoing,
		HTTPPair:    string(httpPair),
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
	mimeType, _ := postData["mimeType"]
	if mimeType == nil || len(mimeType.(string)) == 0 {
		mimeType = "text/html"
	}
	text, _ := postData["text"]
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

func representResponse(response map[string]interface{}) (repResponse []interface{}, bodySize int64) {
	repResponse = make([]interface{}, 0)

	bodySize = int64(response["bodySize"].(float64))

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
			Value:    bodySize,
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
	mimeType, _ := content["mimeType"]
	if mimeType == nil || len(mimeType.(string)) == 0 {
		mimeType = "text/html"
	}
	encoding, _ := content["encoding"]
	text, _ := content["text"]
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

func (d dissecting) Represent(request map[string]interface{}, response map[string]interface{}) (object []byte, bodySize int64, err error) {
	representation := make(map[string]interface{}, 0)
	repRequest := representRequest(request)
	repResponse, bodySize := representResponse(response)
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
	}
}

var Dissector dissecting

func NewDissector() api.Dissector {
	return Dissector
}
