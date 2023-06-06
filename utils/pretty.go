package utils

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

const (
	empty = ""
	tab   = "\t"
)

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
