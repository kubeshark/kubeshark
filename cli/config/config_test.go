package config_test

import (
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
	"testing"
)

func TestConfigWriteIgnoresReadonlyFields(t *testing.T) {
	var readonlyFields []string

	configElem := reflect.ValueOf(&config.ConfigStruct{}).Elem()
	getFieldsWithReadonlyTag(configElem, &readonlyFields)

	configWithDefaults, _ := config.GetConfigWithDefaults()
	configWithDefaultsBytes, _ := yaml.Marshal(configWithDefaults)
	for _, readonlyField := range readonlyFields {
		t.Run(readonlyField, func(t *testing.T) {
			readonlyFieldToCheck := fmt.Sprintf(" %s:", readonlyField)
			if strings.Contains(string(configWithDefaultsBytes), readonlyFieldToCheck) {
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
