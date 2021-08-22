package main

import (
	"encoding/json"
)

type KafkaPayload struct {
	Type string
	Data interface{}
}

type KafkaPayloader interface {
	MarshalJSON() ([]byte, error)
}

func (h KafkaPayload) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Data)
}
