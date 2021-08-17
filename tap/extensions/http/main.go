package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/up9inc/mizu/tap/api"
)

func init() {
	log.Println("Initializing HTTP extension.")
}

var discardBuffer = make([]byte, 4096)

type dissecting string

func (g dissecting) Register(extension *api.Extension) {
	extension.Name = "http"
	extension.OutboundPorts = []string{"80", "8080", "443"}
	extension.InboundPorts = []string{}
}

func (g dissecting) Ping() {
	log.Printf("pong HTTP\n")
}

func (g dissecting) Dissect(b *bufio.Reader, isClient bool) interface{} {
	for {
		if isClient {
			req, err := http.ReadRequest(b)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// We must read until we see an EOF... very important!
				return nil
			} else if err != nil {
				log.Println("Error reading stream:", err)
			} else {
				body, _ := ioutil.ReadAll(req.Body)
				req.Body.Close()
				log.Printf("Received request: %+v with body: %+v\n", req, body)
			}
		} else {
			res, err := http.ReadResponse(b, nil)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// We must read until we see an EOF... very important!
				return nil
			} else if err != nil {
				log.Println("Error reading stream:", err)
			} else {
				body, _ := ioutil.ReadAll(res.Body)
				res.Body.Close()
				log.Printf("Received response: %+v with body: %+v\n", res, body)
			}
		}
	}
}

// exported as symbol named "Greeter"
var Dissector dissecting
