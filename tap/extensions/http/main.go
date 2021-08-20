package main

import (
	"bufio"
	"fmt"
	"io"
	"log"

	"github.com/romana/rlog"

	"github.com/up9inc/mizu/tap/api"
)

var requestCounter uint
var responseCounter uint

func init() {
	log.Println("Initializing HTTP extension.")
	requestCounter = 0
	responseCounter = 0
}

type dissecting string

const ExtensionName = "http"

func (d dissecting) Register(extension *api.Extension) {
	extension.Name = ExtensionName
	extension.OutboundPorts = []string{"80", "8080", "443"}
	extension.InboundPorts = []string{}
}

func (d dissecting) Ping() {
	log.Printf("pong HTTP\n")
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

func (d dissecting) Summarize(item *api.OutputChannelItem) *api.BaseEntryDetails {
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
	return &api.BaseEntryDetails{
		Url:             fmt.Sprintf("http://%s%s", host, request["url"].(string)),
		RequestSenderIp: item.ConnectionInfo.ClientIP,
		Path:            request["url"].(string),
		StatusCode:      int(response["status"].(float64)),
		Method:          request["method"].(string),
		Timestamp:       item.Timestamp,
	}
}

var Dissector dissecting
