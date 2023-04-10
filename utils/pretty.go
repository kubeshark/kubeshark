package utils

import (
	"bytes"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

const (
	empty = ""
	tab   = "\t"
)

func PrettyJson(data interface{}) (string, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(empty, tab)

	err := encoder.Encode(data)
	if err != nil {
		return empty, err
	}
	return buffer.String(), nil
}

func PrettyYaml(data interface{}) (string, error) {
	buffer := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buffer)
	encoder.SetIndent(2)

	err := encoder.Encode(data)
	if err != nil {
		return empty, err
	}
	return buffer.String(), nil
}

func PrettyYamlOmitEmpty(data interface{}) (string, error) {
	d, err := json.Marshal(data)
	if err != nil {
		return empty, err
	}

	var cleanData map[string]interface{}
	err = json.Unmarshal(d, &cleanData)
	if err != nil {
		return empty, err
	}

	return PrettyYaml(cleanData)
}
