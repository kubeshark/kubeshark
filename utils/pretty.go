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

func PrettyYaml(data interface{}) (result string, err error) {
	var marshalled []byte
	marshalled, err = json.Marshal(data)
	if err != nil {
		return
	}

	var unmarshalled interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return
	}

	buffer := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buffer)
	encoder.SetIndent(2)

	err = encoder.Encode(unmarshalled)
	if err != nil {
		return
	}
	result = buffer.String()
	return
}
