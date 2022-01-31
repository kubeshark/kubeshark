package amqp

import (
	"encoding/json"
)

type AMQPPayload struct {
	Data interface{}
}

type AMQPPayloader interface {
	MarshalJSON() ([]byte, error)
}

func (h AMQPPayload) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Data)
}
