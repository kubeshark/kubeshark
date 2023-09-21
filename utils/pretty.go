package utils

import (
	"bytes"
	"encoding/json"

	"github.com/goccy/go-yaml"
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
	encoder := yaml.NewEncoder(buffer, yaml.Indent(2))

	err = encoder.Encode(unmarshalled)
	if err != nil {
		return
	}
	result = buffer.String()
	return
}
