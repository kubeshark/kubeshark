package redis

import (
	"encoding/json"
	"fmt"

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

func representGeneric(generic map[string]interface{}, selectorPrefix string) (representation []interface{}) {
	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Type",
			Value:    generic["type"].(string),
			Selector: fmt.Sprintf("%stype", selectorPrefix),
		},
		{
			Name:     "Command",
			Value:    generic["command"].(string),
			Selector: fmt.Sprintf("%scommand", selectorPrefix),
		},
		{
			Name:     "Key",
			Value:    generic["key"].(string),
			Selector: fmt.Sprintf("%skey", selectorPrefix),
		},
		{
			Name:     "Keyword",
			Value:    generic["keyword"].(string),
			Selector: fmt.Sprintf("%skeyword", selectorPrefix),
		},
	})
	representation = append(representation, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	representation = append(representation, api.SectionData{
		Type:     api.BODY,
		Title:    "Value",
		Data:     generic["value"].(string),
		Selector: fmt.Sprintf("%svalue", selectorPrefix),
	})

	return
}
