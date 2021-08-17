package main

import (
	"log"
)

func init() {
	log.Println("Initializing HTTP extension.")
}

type greeting string

func (g greeting) Greet() {
	log.Println("Hello from HTTP extension.")
}

// exported as symbol named "Greeter"
var Greeter greeting
