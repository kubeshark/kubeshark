package main

import (
	"encoding/json"
	"fmt"
)

type AMQPPayload struct {
	Type   string
	Method string
	Data   interface{}
}

type AMQPPayloader interface {
	MarshalJSON() ([]byte, error)
}

func (h AMQPPayload) MarshalJSON() ([]byte, error) {
	switch h.Type {
	case "basic_publish":
		return json.Marshal(h.Data)
	default:
		panic(fmt.Sprintf("AMQP payload cannot be marshaled: %s\n", h.Type))
	}
}
