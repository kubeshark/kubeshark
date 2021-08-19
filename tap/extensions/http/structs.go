package main

import (
	"fmt"
)

type HTTPPayload struct {
	Type string
	Data interface{}
}

type HTTPPayloader interface {
	MarshalJSON() ([]byte, error)
}

func (h HTTPPayload) MarshalJSON() ([]byte, error) {
	switch h.Type {
	case "http_request":
		return []byte("{\"val\": \"" + h.Type + "\"}"), nil
	case "http_response":
		return []byte("{\"val\": \"" + h.Type + "\"}"), nil
	default:
		panic(fmt.Sprintf("HTTP payload cannot be marshaled: %s\n", h.Type))
	}
}
