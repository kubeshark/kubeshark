package main

import (
	"fmt"
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
	fmt.Printf("pong Kafka\n")
}

// exported as symbol named "Greeter"
var Dissector dissecting
