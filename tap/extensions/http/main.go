package main

import (
	"fmt"
	"log"

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
	fmt.Printf("pong HTTP\n")
}

// exported as symbol named "Greeter"
var Dissector dissecting
