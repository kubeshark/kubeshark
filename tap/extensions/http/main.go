package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

var protocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol -- HTTP/1.1",
	Abbreviation:    "HTTP",
	Macro:           "http",
	Version:         "1.1",
	BackgroundColor: "#205cf5",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://datatracker.ietf.org/doc/html/rfc2616",
	Ports:           []string{"80", "8080", "50051"},
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
	Ports:           []string{"80", "8080"},
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
	extension.Protocol = &protocol
	extension.MatcherMap = reqResMatcher.openMessagesMap
}

func (d dissecting) Ping() {
	log.Printf("pong %s", protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, superIdentifier *api.SuperIdentifier, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	isHTTP2, err := checkIsHTTP2Connection(b, isClient)

	var http2Assembler *Http2Assembler
	if isHTTP2 {
		prepareHTTP2Connection(b, isClient)
		http2Assembler = createHTTP2Assembler(b)
	}

	dissected := false
	for {
		if superIdentifier.Protocol != nil && superIdentifier.Protocol != &protocol {
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
			err = handleHTTP1ClientStream(b, tcpID, counterPair, superTimer, emitter, options)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				continue
			}
			dissected = true
		} else {
			err = handleHTTP1ServerStream(b, tcpID, counterPair, superTimer, emitter, options)
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
	superIdentifier.Protocol = &protocol
	return nil
}

func SetHostname(address, newHostname string) string {
	replacedUrl, err := url.Parse(address)
	if err != nil {
		return address
	}
	replacedUrl.Host = newHostname
	return replacedUrl.String()
}

func (d dissecting) Analyze(item *api.OutputChannelItem, resolvedSource string, resolvedDestination string) *api.MizuEntry {
	var host, scheme, authority, path, service string

	request := item.Pair.Request.Payload.(map[string]interface{})
	response := item.Pair.Response.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	resDetails := response["details"].(map[string]interface{})

	for _, header := range reqDetails["headers"].([]interface{}) {
		h := header.(map[string]interface{})
		if h["name"] == "Host" {
			host = h["value"].(string)
		}
		if h["name"] == ":authority" {
			authority = h["value"].(string)
		}
		if h["name"] == ":scheme" {
			scheme = h["value"].(string)
		}
		if h["name"] == ":path" {
			path = h["value"].(string)
		}
	}

	if resDetails["bodySize"].(float64) < 0 {
		resDetails["bodySize"] = 0
	}

	if item.Protocol.Version == "2.0" {
		service = fmt.Sprintf("%s://%s", scheme, authority)
	} else {
		service = fmt.Sprintf("http://%s", host)
		u, err := url.Parse(reqDetails["url"].(string))
		if err != nil {
			path = reqDetails["url"].(string)
		} else {
			path = u.Path
		}
	}

	request["url"] = reqDetails["url"].(string)
	reqDetails["path"] = path
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
	reqDetails["queryString"] = mapSliceRebuildAsMap(reqDetails["_queryString"].([]interface{}))

	if resolvedDestination != "" {
		service = SetHostname(service, resolvedDestination)
	} else if resolvedSource != "" {
		service = SetHostname(service, resolvedSource)
	}

	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	if elapsedTime < 0 {
		elapsedTime = 0
	}
	httpPair, _ := json.Marshal(item.Pair)
	_protocol := protocol
	_protocol.Version = item.Protocol.Version
	return &api.MizuEntry{
		Protocol: _protocol,
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
		Outgoing:            item.ConnectionInfo.IsOutgoing,
		Request:             reqDetails,
		Response:            resDetails,
		Url:                 fmt.Sprintf("%s%s", service, path),
		Method:              reqDetails["method"].(string),
		Status:              int(resDetails["status"].(float64)),
		RequestSenderIp:     item.ConnectionInfo.ClientIP,
		Service:             service,
		Timestamp:           item.Timestamp,
		StartTime:           item.Pair.Request.CaptureTime,
		ElapsedTime:         elapsedTime,
		Summary:             path,
		ResolvedSource:      resolvedSource,
		ResolvedDestination: resolvedDestination,
		SourceIp:            item.ConnectionInfo.ClientIP,
		DestinationIp:       item.ConnectionInfo.ServerIP,
		SourcePort:          item.ConnectionInfo.ClientPort,
		DestinationPort:     item.ConnectionInfo.ServerPort,
		IsOutgoing:          item.ConnectionInfo.IsOutgoing,
		HTTPPair:            string(httpPair),
	}
}

func (d dissecting) Summarize(entry *api.MizuEntry) *api.BaseEntryDetails {
	var p api.Protocol
	if entry.Protocol.Version == "2.0" {
		p = http2Protocol
	} else {
		p = protocol
	}
	return &api.BaseEntryDetails{
		Id:              entry.Id,
		Protocol:        p,
		Url:             entry.Url,
		RequestSenderIp: entry.RequestSenderIp,
		Service:         entry.Service,
		Path:            entry.Path,
		Summary:         entry.Summary,
		StatusCode:      entry.Status,
		Method:          entry.Method,
		Timestamp:       entry.Timestamp,
		SourceIp:        entry.SourceIp,
		DestinationIp:   entry.DestinationIp,
		SourcePort:      entry.SourcePort,
		DestinationPort: entry.DestinationPort,
		IsOutgoing:      entry.IsOutgoing,
		Latency:         entry.ElapsedTime,
		Rules: api.ApplicableRules{
			Latency: 0,
			Status:  false,
		},
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
			Name:     "URL",
			Value:    request["url"].(string),
			Selector: `request.url`,
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
		Data:  representMapSliceAsTable(request["_queryString"].([]interface{}), `request.queryString`),
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

func (d dissecting) Represent(protoIn api.Protocol, request map[string]interface{}, response map[string]interface{}) (protoOut api.Protocol, object []byte, bodySize int64, err error) {
	if protoIn.Version == "2.0" {
		protoOut = http2Protocol
	} else {
		protoOut = protocol
	}
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
		`http`:  fmt.Sprintf(`proto.abbr == "%s" and proto.version == "%s"`, protocol.Abbreviation, protocol.Version),
		`http2`: fmt.Sprintf(`proto.abbr == "%s" and proto.version == "%s"`, protocol.Abbreviation, http2Protocol.Version),
	}
}

var Dissector dissecting
