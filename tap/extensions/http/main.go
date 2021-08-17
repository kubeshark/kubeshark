package main

import (
	"bufio"
	"log"
	"net/http"

	"github.com/up9inc/mizu/tap/api"
)

func init() {
	log.Println("Initializing HTTP extension.")
}

type dissecting string

func (g dissecting) Register(extension *api.Extension) {
	extension.Name = "http"
	extension.Ports = []string{"80", "8080", "443"}
}

func (g dissecting) Ping() {
	log.Printf("pong HTTP\n")
}

func (g dissecting) Dissect(b *bufio.Reader) interface{} {
	log.Printf("called Dissect!")
	req, _ := http.ReadRequest(b)
	log.Printf("HTTP Request: %+v\n", req)
	return nil
}

// exported as symbol named "Greeter"
var Dissector dissecting
