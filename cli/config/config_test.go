package config_test

import (
	config3 "github.com/up9inc/mizu/cli/config"
	"reflect"
	"strings"
	"testing"
)

func TestConfigWriteIgnoresReadonlyFields(t *testing.T) {
	var readonlyFields []string

	configElem := reflect.ValueOf(&config3.ConfigStruct{}).Elem()
	getFieldsWithReadonlyTag(configElem, &readonlyFields)

	config, _ := config3.GetConfigWithDefaults()
	for _, readonlyField := range readonlyFields {
		if strings.Contains(config, readonlyField) {
			t.Errorf("unexpected result - readonly field: %v, config: %v", readonlyField, config)
		}
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

		if _, ok := currentField.Tag.Lookup(config3.ReadonlyTag); ok {
			fieldNameByTag := strings.Split(currentField.Tag.Get(config3.FieldNameTag), ",")[0]
			*readonlyFields = append(*readonlyFields, fieldNameByTag)
		}
	}
}
