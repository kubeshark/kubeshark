package mizu_test

import (
	"github.com/up9inc/mizu/cli/mizu"
	"reflect"
	"strings"
	"testing"
)

func TestConfigWriteIgnoresReadonlyFields(t *testing.T) {
	var readonlyFields []string

	configElem := reflect.ValueOf(&mizu.ConfigStruct{}).Elem()
	getFieldsWithReadonlyTag(configElem, &readonlyFields)

	config, _ := mizu.GetConfigWithDefaults()
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

		if _, ok := currentField.Tag.Lookup(mizu.ReadonlyTag); ok {
			fieldNameByTag := strings.ReplaceAll(currentField.Tag.Get(mizu.FieldNameTag), ",omitempty", "")
			*readonlyFields = append(*readonlyFields, fieldNameByTag)
		}
	}
}
