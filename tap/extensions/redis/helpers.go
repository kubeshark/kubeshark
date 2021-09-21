package main

import "encoding/json"

type RedisPayload struct {
	Data interface{}
}

type RedisPayloader interface {
	MarshalJSON() ([]byte, error)
}

func (h RedisPayload) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Data)
}

type RedisWrapper struct {
	Method  string      `json:"method"`
	Url     string      `json:"url"`
	Details interface{} `json:"details"`
}
