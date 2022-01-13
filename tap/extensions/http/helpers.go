package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/up9inc/mizu/tap/api"
)

func mapSliceRebuildAsMap(mapSlice []interface{}) (newMap map[string]interface{}) {
	newMap = make(map[string]interface{})
	for _, item := range mapSlice {
		h := item.(map[string]interface{})
		newMap[h["name"].(string)] = h["value"]
	}

	return
}

func representMapSliceAsTable(mapSlice []interface{}, selectorPrefix string) (representation string) {
	var table []api.TableData
	for _, item := range mapSlice {
		h := item.(map[string]interface{})
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

func representSliceAsTable(slice []interface{}, selectorPrefix string) (representation string) {
	var table []api.TableData
	for i, item := range slice {
		selector := fmt.Sprintf("%s[%d]", selectorPrefix, i)
		table = append(table, api.TableData{
			Name:     strconv.Itoa(i),
			Value:    item.(interface{}),
			Selector: selector,
		})
	}

	obj, _ := json.Marshal(table)
	representation = string(obj)
	return
}
