package main

import (
	"bufio"
	"io"
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
	extension.Ports = []string{"80", "8080", "443"}
}

func (g dissecting) Ping() {
	log.Printf("pong HTTP\n")
}

func DiscardBytesToFirstError(r io.Reader) (discarded int, err error) {
	for {
		n, e := r.Read(discardBuffer)
		discarded += n
		if e != nil {
			return discarded, e
		}
	}
}

func DiscardBytesToEOF(r io.Reader) (discarded int) {
	for {
		n, e := DiscardBytesToFirstError(r)
		discarded += n
		if e == io.EOF {
			return
		}
	}
}

func (g dissecting) Dissect(b *bufio.Reader) interface{} {
	for {
		req, err := http.ReadRequest(b)
		if err == io.EOF {
			// We must read until we see an EOF... very important!
			return nil
		} else if err != nil {
			log.Println("Error reading stream:", err)
		} else {
			bodyBytes := DiscardBytesToEOF(req.Body)
			req.Body.Close()
			log.Println("Received request from stream:", req, "with", bodyBytes, "bytes in request body")
		}
	}
}

// exported as symbol named "Greeter"
var Dissector dissecting
