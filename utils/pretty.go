package utils

import (
	"bytes"

	"github.com/goccy/go-yaml"
)

func PrettyYaml(data interface{}) (result string, err error) {
	buffer := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buffer, yaml.Indent(2))

	err = encoder.Encode(data)
	if err != nil {
		return
	}
	result = buffer.String()
	return
}
