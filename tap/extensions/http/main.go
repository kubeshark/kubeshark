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

	"github.com/romana/rlog"

	"github.com/up9inc/mizu/tap/api"
)

var protocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol -- HTTP/1.1",
	Abbreviation:    "HTTP",
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
	LongName:        "Hypertext Transfer Protocol Version 2 (HTTP/2) (gRPC)",
	Abbreviation:    "HTTP/2",
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
	log.Printf("pong %s\n", protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, superIdentifier *api.SuperIdentifier, emitter api.Emitter) error {
	ident := fmt.Sprintf("%s->%s:%s->%s", tcpID.SrcIP, tcpID.DstIP, tcpID.SrcPort, tcpID.DstPort)
	isHTTP2, err := checkIsHTTP2Connection(b, isClient)
	if err != nil {
		rlog.Debugf("[HTTP/2-Prepare-Connection] stream %s Failed to check if client is HTTP/2: %s (%v,%+v)", ident, err, err, err)
		// Do something?
	}

	var grpcAssembler *GrpcAssembler
	if isHTTP2 {
		err := prepareHTTP2Connection(b, isClient)
		if err != nil {
			rlog.Debugf("[HTTP/2-Prepare-Connection-After-Check] stream %s error: %s (%v,%+v)", ident, err, err, err)
		}
		grpcAssembler = createGrpcAssembler(b)
	}

	dissected := false
	for {
		if superIdentifier.Protocol != nil && superIdentifier.Protocol != &protocol {
			return errors.New("Identified by another protocol")
		}

		if isHTTP2 {
			err = handleHTTP2Stream(grpcAssembler, tcpID, superTimer, emitter)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				rlog.Debugf("[HTTP/2] stream %s error: %s (%v,%+v)", ident, err, err, err)
				continue
			}
			dissected = true
		} else if isClient {
			err = handleHTTP1ClientStream(b, tcpID, counterPair, superTimer, emitter)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				rlog.Debugf("[HTTP-request] stream %s Request error: %s (%v,%+v)", ident, err, err, err)
				continue
			}
			dissected = true
		} else {
			err = handleHTTP1ServerStream(b, tcpID, counterPair, superTimer, emitter)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				rlog.Debugf("[HTTP-response], stream %s Response error: %s (%v,%+v)", ident, err, err, err)
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
		log.Printf("error replacing hostname to %s in address %s, returning original %v", newHostname, address, err)
		return address
	}
	replacedUrl.Host = newHostname
	return replacedUrl.String()
}

func (d dissecting) Analyze(item *api.OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *api.MizuEntry {
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

	if item.Protocol.Version == "2.0" {
		service = fmt.Sprintf("%s://%s", scheme, authority)
	} else {
		service = fmt.Sprintf("http://%s", host)
		path = reqDetails["url"].(string)
	}

	request["url"] = path
	if resolvedDestination != "" {
		service = SetHostname(service, resolvedDestination)
	} else if resolvedSource != "" {
		service = SetHostname(service, resolvedSource)
	}

	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	entryBytes, _ := json.Marshal(item.Pair)
	return &api.MizuEntry{
		ProtocolName:            protocol.Name,
		ProtocolLongName:        protocol.LongName,
		ProtocolAbbreviation:    protocol.Abbreviation,
		ProtocolVersion:         item.Protocol.Version,
		ProtocolBackgroundColor: protocol.BackgroundColor,
		ProtocolForegroundColor: protocol.ForegroundColor,
		ProtocolFontSize:        protocol.FontSize,
		ProtocolReferenceLink:   protocol.ReferenceLink,
		EntryId:                 entryId,
		Entry:                   string(entryBytes),
		Url:                     fmt.Sprintf("%s%s", service, path),
		Method:                  reqDetails["method"].(string),
		Status:                  int(resDetails["status"].(float64)),
		RequestSenderIp:         item.ConnectionInfo.ClientIP,
		Service:                 service,
		Timestamp:               item.Timestamp,
		ElapsedTime:             elapsedTime,
		Path:                    path,
		ResolvedSource:          resolvedSource,
		ResolvedDestination:     resolvedDestination,
		SourceIp:                item.ConnectionInfo.ClientIP,
		DestinationIp:           item.ConnectionInfo.ServerIP,
		SourcePort:              item.ConnectionInfo.ClientPort,
		DestinationPort:         item.ConnectionInfo.ServerPort,
		IsOutgoing:              item.ConnectionInfo.IsOutgoing,
	}
}

func (d dissecting) Summarize(entry *api.MizuEntry) *api.BaseEntryDetails {
	var p api.Protocol
	if entry.ProtocolVersion == "2.0" {
		p = http2Protocol
	} else {
		p = protocol
	}
	return &api.BaseEntryDetails{
		Id:              entry.EntryId,
		Protocol:        p,
		Url:             entry.Url,
		RequestSenderIp: entry.RequestSenderIp,
		Service:         entry.Service,
		Path:            entry.Path,
		Summary:         entry.Path,
		StatusCode:      entry.Status,
		Method:          entry.Method,
		Timestamp:       entry.Timestamp,
		SourceIp:        entry.SourceIp,
		DestinationIp:   entry.DestinationIp,
		SourcePort:      entry.SourcePort,
		DestinationPort: entry.DestinationPort,
		IsOutgoing:      entry.IsOutgoing,
		Latency:         0,
		Rules: api.ApplicableRules{
			Latency: 0,
			Status:  false,
		},
	}
}

func representRequest(request map[string]interface{}) (repRequest []interface{}) {
	details, _ := json.Marshal([]map[string]string{
		{
			"name":  "Method",
			"value": request["method"].(string),
		},
		{
			"name":  "URL",
			"value": request["url"].(string),
		},
		{
			"name":  "Body Size",
			"value": fmt.Sprintf("%g bytes", request["bodySize"].(float64)),
		},
	})
	repRequest = append(repRequest, map[string]string{
		"type":  api.TABLE,
		"title": "Details",
		"data":  string(details),
	})

	headers, _ := json.Marshal(request["headers"].([]interface{}))
	repRequest = append(repRequest, map[string]string{
		"type":  api.TABLE,
		"title": "Headers",
		"data":  string(headers),
	})

	cookies, _ := json.Marshal(request["cookies"].([]interface{}))
	repRequest = append(repRequest, map[string]string{
		"type":  api.TABLE,
		"title": "Cookies",
		"data":  string(cookies),
	})

	queryString, _ := json.Marshal(request["queryString"].([]interface{}))
	repRequest = append(repRequest, map[string]string{
		"type":  api.TABLE,
		"title": "Query String",
		"data":  string(queryString),
	})

	postData, _ := request["postData"].(map[string]interface{})
	mimeType, _ := postData["mimeType"]
	if mimeType == nil || len(mimeType.(string)) == 0 {
		mimeType = "text/html"
	}
	text, _ := postData["text"]
	if text != nil {
		repRequest = append(repRequest, map[string]string{
			"type":      api.BODY,
			"title":     "POST Data (text/plain)",
			"encoding":  "",
			"mime_type": mimeType.(string),
			"data":      text.(string),
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
				repRequest = append(repRequest, map[string]string{
					"type":  api.TABLE,
					"title": "POST Data (multipart/form-data)",
					"data":  string(multipart),
				})
			} else {
				repRequest = append(repRequest, map[string]string{
					"type":  api.TABLE,
					"title": "POST Data (application/x-www-form-urlencoded)",
					"data":  string(params),
				})
			}
		}
	}

	return
}

func representResponse(response map[string]interface{}) (repResponse []interface{}, bodySize int64) {
	repResponse = make([]interface{}, 0)

	bodySize = int64(response["bodySize"].(float64))

	details, _ := json.Marshal([]map[string]string{
		{
			"name":  "Status",
			"value": fmt.Sprintf("%g", response["status"].(float64)),
		},
		{
			"name":  "Status Text",
			"value": response["statusText"].(string),
		},
		{
			"name":  "Body Size",
			"value": fmt.Sprintf("%d bytes", bodySize),
		},
	})
	repResponse = append(repResponse, map[string]string{
		"type":  api.TABLE,
		"title": "Details",
		"data":  string(details),
	})

	headers, _ := json.Marshal(response["headers"].([]interface{}))
	repResponse = append(repResponse, map[string]string{
		"type":  api.TABLE,
		"title": "Headers",
		"data":  string(headers),
	})

	cookies, _ := json.Marshal(response["cookies"].([]interface{}))
	repResponse = append(repResponse, map[string]string{
		"type":  api.TABLE,
		"title": "Cookies",
		"data":  string(cookies),
	})

	content, _ := response["content"].(map[string]interface{})
	mimeType, _ := content["mimeType"]
	if mimeType == nil || len(mimeType.(string)) == 0 {
		mimeType = "text/html"
	}
	encoding, _ := content["encoding"]
	text, _ := content["text"]
	if text != nil {
		repResponse = append(repResponse, map[string]string{
			"type":      api.BODY,
			"title":     "Body",
			"encoding":  encoding.(string),
			"mime_type": mimeType.(string),
			"data":      text.(string),
		})
	}

	return
}

func (d dissecting) Represent(entry *api.MizuEntry) (p api.Protocol, object []byte, bodySize int64, err error) {
	if entry.ProtocolVersion == "2.0" {
		p = http2Protocol
	} else {
		p = protocol
	}
	var root map[string]interface{}
	json.Unmarshal([]byte(entry.Entry), &root)
	representation := make(map[string]interface{}, 0)
	request := root["request"].(map[string]interface{})["payload"].(map[string]interface{})
	response := root["response"].(map[string]interface{})["payload"].(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	resDetails := response["details"].(map[string]interface{})
	repRequest := representRequest(reqDetails)
	repResponse, bodySize := representResponse(resDetails)
	representation["request"] = repRequest
	representation["response"] = repResponse
	object, err = json.Marshal(representation)
	return
}

var Dissector dissecting
