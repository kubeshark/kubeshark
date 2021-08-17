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
	extension.Ports = []string{"9092"}
}

func (g dissecting) Ping() {
	log.Printf("pong Kafka\n")
}

func (g dissecting) Dissect(b *bufio.Reader) interface{} {
	// TODO: Implement
	return nil
}

// exported as symbol named "Greeter"
var Dissector dissecting
