package config_test

import (
	"github.com/up9inc/mizu/cli/config"
	"reflect"
	"strings"
	"testing"
)

func TestConfigWriteIgnoresReadonlyFields(t *testing.T) {
	var readonlyFields []string

	configElem := reflect.ValueOf(&config.ConfigStruct{}).Elem()
	getFieldsWithReadonlyTag(configElem, &readonlyFields)

	configWithDefaults, _ := config.GetConfigWithDefaults()
	for _, readonlyField := range readonlyFields {
		t.Run(readonlyField, func(t *testing.T) {
			if strings.Contains(configWithDefaults, readonlyField) {
				t.Errorf("unexpected result - readonly field: %v, config: %v", readonlyField, configWithDefaults)
			}
		})
	}
}

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
