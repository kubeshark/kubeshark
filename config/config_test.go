package config_test

import (
	"reflect"
	"strings"

	"github.com/kubeshark/kubeshark/config"
)

func getFieldsWithReadonlyTag(currentElem reflect.Value, readonlyFields *[]string) {
	for i := 0; i < currentElem.NumField(); i++ {
		currentField := currentElem.Type().Field(i)
		currentFieldByName := currentElem.FieldByName(currentField.Name)

		if currentField.Type.Kind() == reflect.Struct {
			getFieldsWithReadonlyTag(currentFieldByName, readonlyFields)
			continue
		}

		if _, ok := currentField.Tag.Lookup(config.ReadonlyTag); ok {
			fieldNameByTag := strings.Split(currentField.Tag.Get(config.FieldNameTag), ",")[0]
			*readonlyFields = append(*readonlyFields, fieldNameByTag)
		}
	}
}
