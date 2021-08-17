package main

import (
	"log"
)

func init() {
	log.Println("Initializing Kafka extension.")
}

type greeting string

func (g greeting) Greet() {
	log.Println("Hello from Kafka extension.")
}

// exported as symbol named "Greeter"
var Greeter greeting
