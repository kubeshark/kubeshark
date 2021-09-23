package main

import (
	"encoding/json"

	"github.com/up9inc/mizu/tap/api"
)

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

func representGeneric(generic map[string]string) (representation []interface{}) {
	details, _ := json.Marshal([]map[string]string{
		{
			"name":  "Type",
			"value": generic["type"],
		},
		{
			"name":  "Command",
			"value": generic["command"],
		},
		{
			"name":  "Key",
			"value": generic["key"],
		},
		{
			"name":  "Value",
			"value": generic["value"],
		},
		{
			"name":  "Keyword",
			"value": generic["keyword"],
		},
	})
	representation = append(representation, map[string]string{
		"type":  api.TABLE,
		"title": "Details",
		"data":  string(details),
	})

	return
}
