package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/romana/rlog"

	"github.com/up9inc/mizu/tap/api"
)

var requestCounter uint
var responseCounter uint

var protocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol -- HTTP/1.0",
	Abbreviation:    "HTTP",
	BackgroundColor: "#205cf5",
	ForegroundColor: "#ffffff",
	FontSize:        10,
	ReferenceLink:   "https://www.ietf.org/rfc/rfc1945.txt",
	OutboundPorts:   []string{"80", "8080", "443"},
	InboundPorts:    []string{},
}

var http2Protocol api.Protocol = api.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol Version 2 (HTTP/2)",
	Abbreviation:    "HTTP/2",
	BackgroundColor: "#244c5a",
	ForegroundColor: "#ffffff",
	FontSize:        10,
	ReferenceLink:   "https://datatracker.ietf.org/doc/html/rfc7540",
	OutboundPorts:   []string{"80", "8080", "443"},
	InboundPorts:    []string{},
}

func init() {
	log.Println("Initializing HTTP extension.")
	requestCounter = 0
	responseCounter = 0
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s\n", protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, emitter api.Emitter) {
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

	for {
		if isHTTP2 {
			err = handleHTTP2Stream(grpcAssembler, tcpID, emitter)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				rlog.Debugf("[HTTP/2] stream %s error: %s (%v,%+v)", ident, err, err, err)
				continue
			}
		} else if isClient {
			err = handleHTTP1ClientStream(b, tcpID, emitter)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				rlog.Debugf("[HTTP-request] stream %s Request error: %s (%v,%+v)", ident, err, err, err)
				continue
			}
		} else {
			err = handleHTTP1ServerStream(b, tcpID, emitter)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				rlog.Debugf("[HTTP-response], stream %s Response error: %s (%v,%+v)", ident, err, err, err)
				continue
			}
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *api.MizuEntry {
	fmt.Printf("pair.Request.Payload: %+v\n", item.Pair.Request.Payload)
	fmt.Printf("item.Pair.Response.Payload: %+v\n", item.Pair.Response.Payload)
	var host string
	for _, header := range item.Pair.Request.Payload.(map[string]interface{})["headers"].([]interface{}) {
		h := header.(map[string]interface{})
		if h["name"] == "Host" {
			host = h["value"].(string)
		}
	}
	request := item.Pair.Request.Payload.(map[string]interface{})
	response := item.Pair.Response.Payload.(map[string]interface{})
	entryBytes, _ := json.Marshal(item.Pair)
	service := fmt.Sprintf("http://%s", host)
	return &api.MizuEntry{
		EntryId:             entryId,
		Entry:               string(entryBytes),
		Url:                 fmt.Sprintf("%s%s", service, request["url"].(string)),
		Method:              request["method"].(string),
		Status:              int(response["status"].(float64)),
		RequestSenderIp:     item.ConnectionInfo.ClientIP,
		Service:             service,
		Timestamp:           item.Timestamp,
		Path:                request["url"].(string),
		ResolvedSource:      resolvedSource,
		ResolvedDestination: resolvedDestination,
		IsOutgoing:          item.ConnectionInfo.IsOutgoing,
	}
}

func (d dissecting) Summarize(entry *api.MizuEntry) *api.BaseEntryDetails {
	return &api.BaseEntryDetails{
		Id:              entry.EntryId,
		Protocol:        protocol,
		Url:             entry.Url,
		RequestSenderIp: entry.RequestSenderIp,
		Service:         entry.Service,
		Path:            entry.Path,
		StatusCode:      entry.Status,
		Method:          entry.Method,
		Timestamp:       entry.Timestamp,
		IsOutgoing:      entry.IsOutgoing,
		Latency:         0,
		Rules: api.ApplicableRules{
			Latency: 0,
			Status:  false,
		},
	}
}

var Dissector dissecting
