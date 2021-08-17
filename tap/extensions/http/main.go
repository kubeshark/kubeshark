package main

import (
	"fmt"

	"github.com/up9inc/mizu/tap/extensions/http/lib"
)

type greeting string

func (g greeting) Greet() {
	fmt.Println("Hello HTTP")
	lib.SomeExample()
}

// exported as symbol named "Greeter"
var Greeter greeting
