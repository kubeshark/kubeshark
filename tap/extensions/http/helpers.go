package main

import (
	"encoding/json"
	"fmt"

	"github.com/up9inc/mizu/tap/api"
)

func mapSliceRebuildAsMap(mapSlice []interface{}) (newMap map[string]interface{}) {
	newMap = make(map[string]interface{})
	for _, header := range mapSlice {
		h := header.(map[string]interface{})
		newMap[h["name"].(string)] = h["value"]
	}

	return
}

func representMapSliceAsTable(mapSlice []interface{}, selectorPrefix string) (representation string) {
	var table []api.TableData
	for _, header := range mapSlice {
		h := header.(map[string]interface{})
		selector := fmt.Sprintf("%s[\"%s\"]", selectorPrefix, h["name"].(string))
		table = append(table, api.TableData{
			Name:     h["name"].(string),
			Value:    h["value"],
			Selector: selector,
		})
	}

	obj, _ := json.Marshal(table)
	representation = string(obj)
	return
}
