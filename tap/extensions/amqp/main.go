package main

import (
	"fmt"
	"log"

	"github.com/up9inc/mizu/tap/api"
)

func init() {
	log.Println("Initializing AMQP extension.")
}

type dissecting string

func (g dissecting) Register(extension *api.Extension) {
	extension.Name = "amqp"
	extension.Ports = []string{"5671", "5672"}
}

func (g dissecting) Ping() {
	fmt.Printf("pong AMQP\n")
}

// exported as symbol named "Greeter"
var Dissector dissecting
