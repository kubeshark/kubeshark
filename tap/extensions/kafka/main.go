package main

import (
	"bufio"
	"log"

	"github.com/up9inc/mizu/tap/api"
)

func init() {
	log.Println("Initializing Kafka extension.")
}

type dissecting string

func (g dissecting) Register(extension *api.Extension) {
	extension.Name = "kafka"
	extension.OutboundPorts = []string{"9092"}
	extension.InboundPorts = []string{}
}

func (g dissecting) Ping() {
	log.Printf("pong Kafka\n")
}

func (g dissecting) Dissect(b *bufio.Reader, isClient bool) interface{} {
	// TODO: Implement
	return nil
}

// exported as symbol named "Greeter"
var Dissector dissecting
