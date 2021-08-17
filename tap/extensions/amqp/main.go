package main

import (
	"log"
)

func init() {
	log.Println("Initializing AMQP extension.")
}

type greeting string

func (g greeting) Greet() {
	log.Println("Hello from AMQP extension.")
}

// exported as symbol named "Greeter"
var Greeter greeting
