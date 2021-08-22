package main

import (
	"encoding/json"
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
	return json.Marshal(h.Data)
	// switch h.Type {
	// case "amqp_request":
	// 	return json.Marshal(h.Data)
	// default:
	// 	panic(fmt.Sprintf("AMQP payload cannot be marshaled: %s\n", h.Type))
	// }
}
