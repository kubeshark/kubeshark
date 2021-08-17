package main

import (
	"fmt"
)

type greeting string

func (g greeting) Greet() {
	fmt.Println("Hello HTTP")
}

// exported as symbol named "Greeter"
var Greeter greeting
