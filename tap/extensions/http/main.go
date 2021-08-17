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
	fmt.Printf("extension: %v\n", extension)
	extension.Port = 80
}

// exported as symbol named "Greeter"
var Dissector dissecting
