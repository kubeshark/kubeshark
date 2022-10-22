package http

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kubeshark/kubeshark/tap/api"
)

func mapSliceRebuildAsMap(mapSlice []interface{}) (newMap map[string]interface{}) {
	newMap = make(map[string]interface{})

	mergedMapSlice := mapSliceMergeRepeatedKeys(mapSlice)
	for _, item := range mergedMapSlice {
		h := item.(map[string]interface{})
		newMap[h["name"].(string)] = h["value"]
	}

	return
}

func mapSliceRebuildAsMergedMap(mapSlice []interface{}) (newMap map[string]interface{}) {
	newMap = make(map[string]interface{})

	mergedMapSlice := mapSliceMergeRepeatedKeys(mapSlice)
	for _, item := range mergedMapSlice {
		h := item.(map[string]interface{})

		if valuesInterface, ok := h["value"].([]interface{}); ok {
			var values []string
			for _, valueInterface := range valuesInterface {
				values = append(values, valueInterface.(string))
			}

			newMap[h["name"].(string)] = strings.Join(values, ",")
		} else {
			newMap[h["name"].(string)] = h["value"]
		}
	}

	return
}

func mapSliceMergeRepeatedKeys(mapSlice []interface{}) (newMapSlice []interface{}) {
	newMapSlice = make([]interface{}, 0)
	valuesMap := make(map[string][]interface{})
	for _, item := range mapSlice {
		h := item.(map[string]interface{})
		key := h["name"].(string)
		valuesMap[key] = append(valuesMap[key], h["value"])
	}

	for key, values := range valuesMap {
		h := make(map[string]interface{})
		h["name"] = key
		if len(values) == 1 {
			h["value"] = values[0]
		} else {
			h["value"] = values
		}
		newMapSlice = append(newMapSlice, h)
	}

	sort.Slice(newMapSlice, func(i, j int) bool {
		return newMapSlice[i].(map[string]interface{})["name"].(string) < newMapSlice[j].(map[string]interface{})["name"].(string)
	})

	return
}

func representMapAsTable(mapToTable map[string]interface{}, selectorPrefix string) (representation string) {
	var table []api.TableData

	keys := make([]string, 0, len(mapToTable))
	for k := range mapToTable {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		table = append(table, createTableForKey(key, mapToTable[key], selectorPrefix)...)
	}

	obj, _ := json.Marshal(table)
	representation = string(obj)
	return
}

func representMapSliceAsTable(mapSlice []interface{}, selectorPrefix string) (representation string) {
	var table []api.TableData
	for _, item := range mapSlice {
		h := item.(map[string]interface{})
		key := h["name"].(string)
		value := h["value"]

		table = append(table, createTableForKey(key, value, selectorPrefix)...)
	}

	obj, _ := json.Marshal(table)
	representation = string(obj)
	return
}

func createTableForKey(key string, value interface{}, selectorPrefix string) []api.TableData {
	var table []api.TableData

	var reflectKind reflect.Kind
	reflectType := reflect.TypeOf(value)
	if reflectType == nil {
		reflectKind = reflect.Interface
	} else {
		reflectKind = reflect.TypeOf(value).Kind()
	}

	switch reflectKind {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		for i, el := range value.([]interface{}) {
			selector := fmt.Sprintf("%s.%s[%d]", selectorPrefix, key, i)
			table = append(table, api.TableData{
				Name:     fmt.Sprintf("%s [%d]", key, i),
				Value:    el,
				Selector: selector,
			})
		}
	default:
		selector := fmt.Sprintf("%s[\"%s\"]", selectorPrefix, key)
		table = append(table, api.TableData{
			Name:     key,
			Value:    value,
			Selector: selector,
		})
	}

	return table
}

func representSliceAsTable(slice []interface{}, selectorPrefix string) (representation string) {
	var table []api.TableData
	for i, item := range slice {
		selector := fmt.Sprintf("%s[%d]", selectorPrefix, i)
		table = append(table, api.TableData{
			Name:     strconv.Itoa(i),
			Value:    item,
			Selector: selector,
		})
	}

	obj, _ := json.Marshal(table)
	representation = string(obj)
	return
}
