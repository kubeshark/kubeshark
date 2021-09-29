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

func representGeneric(generic map[string]interface{}) (representation []interface{}) {
	details, _ := json.Marshal([]map[string]string{
		{
			"name":  "Type",
			"value": generic["type"].(string),
		},
		{
			"name":  "Command",
			"value": generic["command"].(string),
		},
		{
			"name":  "Key",
			"value": generic["key"].(string),
		},
		{
			"name":  "Value",
			"value": generic["value"].(string),
		},
		{
			"name":  "Keyword",
			"value": generic["keyword"].(string),
		},
	})
	representation = append(representation, map[string]string{
		"type":  api.TABLE,
		"title": "Details",
		"data":  string(details),
	})

	return
}
