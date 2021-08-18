package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

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

func (g dissecting) Register(extension *api.Extension) {
	extension.Name = ExtensionName
	extension.OutboundPorts = []string{"80", "8080", "443"}
	extension.InboundPorts = []string{}
}

func (g dissecting) Ping() {
	log.Printf("pong HTTP\n")
}

func (g dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID) *api.OutputChannelItem {
	for {
		if isClient {
			requestCounter++
			req, err := http.ReadRequest(b)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			} else if err != nil {
				log.Println("Error reading stream:", err)
			} else {
				body, _ := ioutil.ReadAll(req.Body)
				req.Body.Close()
				log.Printf("Received request: %+v with body: %+v\n", req, body)
			}

			ident := fmt.Sprintf(
				"%s->%s %s->%s %d",
				tcpID.SrcIP,
				tcpID.DstIP,
				tcpID.SrcPort,
				tcpID.DstPort,
				requestCounter,
			)
			reqResMatcher.registerRequest(ident, req, time.Now())
		} else {
			responseCounter++
			res, err := http.ReadResponse(b, nil)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			} else if err != nil {
				log.Println("Error reading stream:", err)
			} else {
				body, _ := ioutil.ReadAll(res.Body)
				res.Body.Close()
				log.Printf("Received response: %+v with body: %+v\n", res, body)
			}
			ident := fmt.Sprintf(
				"%s->%s %s->%s %d",
				tcpID.DstIP,
				tcpID.SrcIP,
				tcpID.DstPort,
				tcpID.SrcPort,
				responseCounter,
			)
			reqResPair := reqResMatcher.registerResponse(ident, res, time.Now())
			if reqResPair != nil {
				return reqResPair
			}
		}
	}
	return nil
}

var Dissector dissecting
