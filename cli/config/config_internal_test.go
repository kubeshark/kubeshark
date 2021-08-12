package config

import (
	"reflect"
	"testing"
)

type ConfigMock struct {
	SectionMock      SectionMock `yaml:"section"`
	Test             string      `yaml:"test"`
	StringField      string      `yaml:"string-field"`
	IntField         int         `yaml:"int-field"`
	BoolField        bool        `yaml:"bool-field"`
	UintField        uint        `yaml:"uint-field"`
	StringSliceField []string    `yaml:"string-slice-field"`
	IntSliceField    []int       `yaml:"int-slice-field"`
	BoolSliceField   []bool      `yaml:"bool-slice-field"`
	UintSliceField   []uint      `yaml:"uint-slice-field"`
}

type SectionMock struct {
	Test string `yaml:"test"`
}

func TestMergeSetFlagNoSeparator(t *testing.T) {
	tests := [][]string{{""}, {"t"}, {"", "t"}, {"t", "test", "test:true"}, {"test", "test:true", "testing!", "true"}}

	for _, setValues := range tests {
		configMock := ConfigMock{}
		configMockElemValue := reflect.ValueOf(&configMock).Elem()

		err := mergeSetFlag(configMockElemValue, setValues)

		if err == nil {
			t.Errorf("unexpected unhandled error - setValues: %v", setValues)
			continue
		}

		for i := 0; i < configMockElemValue.NumField(); i++ {
			currentField := configMockElemValue.Type().Field(i)
			currentFieldByName := configMockElemValue.FieldByName(currentField.Name)

			if !currentFieldByName.IsZero() {
				t.Errorf("unexpected value with not default value - setValues: %v", setValues)
			}
		}
	}
}

func TestMergeSetFlagInvalidFlagName(t *testing.T) {
	tests := [][]string{{"invalid_flag=true"}, {"section.invalid_flag=test"}, {"section=test"}, {"=true"}, {"invalid_flag=true", "config.invalid_flag=test", "section=test", "=true"}}

	for _, setValues := range tests {
		configMock := ConfigMock{}
		configMockElemValue := reflect.ValueOf(&configMock).Elem()

		err := mergeSetFlag(configMockElemValue, setValues)

		if err == nil {
			t.Errorf("unexpected unhandled error - setValues: %v", setValues)
			continue
		}

		for i := 0; i < configMockElemValue.NumField(); i++ {
			currentField := configMockElemValue.Type().Field(i)
			currentFieldByName := configMockElemValue.FieldByName(currentField.Name)

			if !currentFieldByName.IsZero() {
				t.Errorf("unexpected case - setValues: %v", setValues)
			}
		}
	}
}

func TestMergeSetFlagInvalidFlagValue(t *testing.T) {
	tests := [][]string{{"int-field=true"}, {"bool-field:5"}, {"uint-field=-1"}, {"int-slice-field=true"}, {"bool-slice-field=5"}, {"uint-slice-field=-1"}, {"int-field=6", "int-field=66"}}

	for _, setValues := range tests {
		configMock := ConfigMock{}
		configMockElemValue := reflect.ValueOf(&configMock).Elem()

		err := mergeSetFlag(configMockElemValue, setValues)

		if err == nil {
			t.Errorf("unexpected unhandled error - setValues: %v", setValues)
			continue
		}

		for i := 0; i < configMockElemValue.NumField(); i++ {
			currentField := configMockElemValue.Type().Field(i)
			currentFieldByName := configMockElemValue.FieldByName(currentField.Name)

			if !currentFieldByName.IsZero() {
				t.Errorf("unexpected case - setValues: %v", setValues)
			}
		}
	}
}

func TestMergeSetFlagNotSliceValues(t *testing.T) {
	tests := [][]struct {
		SetValue   string
		FieldName  string
		FieldValue interface{}
	}{
		{{SetValue: "string-field=test", FieldName: "StringField", FieldValue: "test"}},
		{{SetValue: "int-field=6", FieldName: "IntField", FieldValue: 6}},
		{{SetValue: "bool-field=true", FieldName: "BoolField", FieldValue: true}},
		{{SetValue: "uint-field=6", FieldName: "UintField", FieldValue: uint(6)}},
		{
			{SetValue: "string-field=test", FieldName: "StringField", FieldValue: "test"},
			{SetValue: "int-field=6", FieldName: "IntField", FieldValue: 6},
			{SetValue: "bool-field=true", FieldName: "BoolField", FieldValue: true},
			{SetValue: "uint-field=6", FieldName: "UintField", FieldValue: uint(6)},
		},
	}

	for _, test := range tests {
		configMock := ConfigMock{}
		configMockElemValue := reflect.ValueOf(&configMock).Elem()

		var setValues []string
		for _, setValueInfo := range test {
			setValues = append(setValues, setValueInfo.SetValue)
		}

		err := mergeSetFlag(configMockElemValue, setValues)

		if err != nil {
			t.Errorf("unexpected error result - err: %v", err)
			continue
		}

		for _, setValueInfo := range test {
			fieldValue := configMockElemValue.FieldByName(setValueInfo.FieldName).Interface()
			if fieldValue != setValueInfo.FieldValue {
				t.Errorf("unexpected result - expected: %v, actual: %v", setValueInfo.FieldValue, fieldValue)
			}
		}
	}
}

func TestMergeSetFlagSliceValues(t *testing.T) {
	tests := [][]struct {
		SetValues  []string
		FieldName  string
		FieldValue interface{}
	}{
		{{SetValues: []string{"string-slice-field=test"}, FieldName: "StringSliceField", FieldValue: []string{"test"}}},
		{{SetValues: []string{"int-slice-field=6"}, FieldName: "IntSliceField", FieldValue: []int{6}}},
		{{SetValues: []string{"bool-slice-field=true"}, FieldName: "BoolSliceField", FieldValue: []bool{true}}},
		{{SetValues: []string{"uint-slice-field=6"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6)}}},
		{
			{SetValues: []string{"string-slice-field=test"}, FieldName: "StringSliceField", FieldValue: []string{"test"}},
			{SetValues: []string{"int-slice-field=6"}, FieldName: "IntSliceField", FieldValue: []int{6}},
			{SetValues: []string{"bool-slice-field=true"}, FieldName: "BoolSliceField", FieldValue: []bool{true}},
			{SetValues: []string{"uint-slice-field=6"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6)}},
		},
		{{SetValues: []string{"string-slice-field=test", "string-slice-field=test2"}, FieldName: "StringSliceField", FieldValue: []string{"test", "test2"}}},
		{{SetValues: []string{"int-slice-field=6", "int-slice-field=66"}, FieldName: "IntSliceField", FieldValue: []int{6, 66}}},
		{{SetValues: []string{"bool-slice-field=true", "bool-slice-field=false"}, FieldName: "BoolSliceField", FieldValue: []bool{true, false}}},
		{{SetValues: []string{"uint-slice-field=6", "uint-slice-field=66"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6), uint(66)}}},
		{
			{SetValues: []string{"string-slice-field=test", "string-slice-field=test2"}, FieldName: "StringSliceField", FieldValue: []string{"test", "test2"}},
			{SetValues: []string{"int-slice-field=6", "int-slice-field=66"}, FieldName: "IntSliceField", FieldValue: []int{6, 66}},
			{SetValues: []string{"bool-slice-field=true", "bool-slice-field=false"}, FieldName: "BoolSliceField", FieldValue: []bool{true, false}},
			{SetValues: []string{"uint-slice-field=6", "uint-slice-field=66"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6), uint(66)}},
		},
	}

	for _, test := range tests {
		configMock := ConfigMock{}
		configMockElemValue := reflect.ValueOf(&configMock).Elem()

		var setValues []string
		for _, setValueInfo := range test {
			for _, setValue := range setValueInfo.SetValues {
				setValues = append(setValues, setValue)
			}
		}

		err := mergeSetFlag(configMockElemValue, setValues)

		if err != nil {
			t.Errorf("unexpected error result - err: %v", err)
			continue
		}

		for _, setValueInfo := range test {
			fieldValue := configMockElemValue.FieldByName(setValueInfo.FieldName).Interface()
			if !reflect.DeepEqual(fieldValue, setValueInfo.FieldValue) {
				t.Errorf("unexpected result - expected: %v, actual: %v", setValueInfo.FieldValue, fieldValue)
			}
		}
	}
}

func TestMergeSetFlagMixValues(t *testing.T) {
	tests := [][]struct {
		SetValues  []string
		FieldName  string
		FieldValue interface{}
	}{
		{
			{SetValues: []string{"string-slice-field=test"}, FieldName: "StringSliceField", FieldValue: []string{"test"}},
			{SetValues: []string{"int-slice-field=6"}, FieldName: "IntSliceField", FieldValue: []int{6}},
			{SetValues: []string{"bool-slice-field=true"}, FieldName: "BoolSliceField", FieldValue: []bool{true}},
			{SetValues: []string{"uint-slice-field=6"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6)}},
			{SetValues: []string{"string-field=test"}, FieldName: "StringField", FieldValue: "test"},
			{SetValues: []string{"int-field=6"}, FieldName: "IntField", FieldValue: 6},
			{SetValues: []string{"bool-field=true"}, FieldName: "BoolField", FieldValue: true},
			{SetValues: []string{"uint-field=6"}, FieldName: "UintField", FieldValue: uint(6)},
		},
		{
			{SetValues: []string{"string-slice-field=test", "string-slice-field=test2"}, FieldName: "StringSliceField", FieldValue: []string{"test", "test2"}},
			{SetValues: []string{"int-slice-field=6", "int-slice-field=66"}, FieldName: "IntSliceField", FieldValue: []int{6, 66}},
			{SetValues: []string{"bool-slice-field=true", "bool-slice-field=false"}, FieldName: "BoolSliceField", FieldValue: []bool{true, false}},
			{SetValues: []string{"uint-slice-field=6", "uint-slice-field=66"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6), uint(66)}},
			{SetValues: []string{"string-field=test"}, FieldName: "StringField", FieldValue: "test"},
			{SetValues: []string{"int-field=6"}, FieldName: "IntField", FieldValue: 6},
			{SetValues: []string{"bool-field=true"}, FieldName: "BoolField", FieldValue: true},
			{SetValues: []string{"uint-field=6"}, FieldName: "UintField", FieldValue: uint(6)},
		},
	}

	for _, test := range tests {
		configMock := ConfigMock{}
		configMockElemValue := reflect.ValueOf(&configMock).Elem()

		var setValues []string
		for _, setValueInfo := range test {
			for _, setValue := range setValueInfo.SetValues {
				setValues = append(setValues, setValue)
			}
		}

		err := mergeSetFlag(configMockElemValue, setValues)

		if err != nil {
			t.Errorf("unexpected error result - err: %v", err)
			continue
		}

		for _, setValueInfo := range test {
			fieldValue := configMockElemValue.FieldByName(setValueInfo.FieldName).Interface()
			if !reflect.DeepEqual(fieldValue, setValueInfo.FieldValue) {
				t.Errorf("unexpected result - expected: %v, actual: %v", setValueInfo.FieldValue, fieldValue)
			}
		}
	}
}

func TestGetParsedValueValidValue(t *testing.T) {
	tests := []struct {
		StringValue string
		Kind        reflect.Kind
		ActualValue interface{}
	}{
		{StringValue: "test", Kind: reflect.String, ActualValue: "test"},
		{StringValue: "123", Kind: reflect.String, ActualValue: "123"},
		{StringValue: "true", Kind: reflect.Bool, ActualValue: true},
		{StringValue: "false", Kind: reflect.Bool, ActualValue: false},
		{StringValue: "6", Kind: reflect.Int, ActualValue: 6},
		{StringValue: "-6", Kind: reflect.Int, ActualValue: -6},
		{StringValue: "6", Kind: reflect.Int8, ActualValue: int8(6)},
		{StringValue: "-6", Kind: reflect.Int8, ActualValue: int8(-6)},
		{StringValue: "6", Kind: reflect.Int16, ActualValue: int16(6)},
		{StringValue: "-6", Kind: reflect.Int16, ActualValue: int16(-6)},
		{StringValue: "6", Kind: reflect.Int32, ActualValue: int32(6)},
		{StringValue: "-6", Kind: reflect.Int32, ActualValue: int32(-6)},
		{StringValue: "6", Kind: reflect.Int64, ActualValue: int64(6)},
		{StringValue: "-6", Kind: reflect.Int64, ActualValue: int64(-6)},
		{StringValue: "6", Kind: reflect.Uint, ActualValue: uint(6)},
		{StringValue: "66", Kind: reflect.Uint, ActualValue: uint(66)},
		{StringValue: "6", Kind: reflect.Uint8, ActualValue: uint8(6)},
		{StringValue: "66", Kind: reflect.Uint8, ActualValue: uint8(66)},
		{StringValue: "6", Kind: reflect.Uint16, ActualValue: uint16(6)},
		{StringValue: "66", Kind: reflect.Uint16, ActualValue: uint16(66)},
		{StringValue: "6", Kind: reflect.Uint32, ActualValue: uint32(6)},
		{StringValue: "66", Kind: reflect.Uint32, ActualValue: uint32(66)},
		{StringValue: "6", Kind: reflect.Uint64, ActualValue: uint64(6)},
		{StringValue: "66", Kind: reflect.Uint64, ActualValue: uint64(66)},
	}

	for _, test := range tests {
		parsedValue, err := getParsedValue(test.Kind, test.StringValue)

		if err != nil {
			t.Errorf("unexpected error result - err: %v", err)
			continue
		}

		if parsedValue.Interface() != test.ActualValue {
			t.Errorf("unexpected result - expected: %v, actual: %v", test.ActualValue, parsedValue)
		}
	}
}

func TestGetParsedValueInvalidValue(t *testing.T) {
	tests := []struct {
		StringValue string
		Kind        reflect.Kind
	}{
		{StringValue: "test", Kind: reflect.Bool},
		{StringValue: "123", Kind: reflect.Bool},
		{StringValue: "test", Kind: reflect.Int},
		{StringValue: "true", Kind: reflect.Int},
		{StringValue: "test", Kind: reflect.Int8},
		{StringValue: "true", Kind: reflect.Int8},
		{StringValue: "test", Kind: reflect.Int16},
		{StringValue: "true", Kind: reflect.Int16},
		{StringValue: "test", Kind: reflect.Int32},
		{StringValue: "true", Kind: reflect.Int32},
		{StringValue: "test", Kind: reflect.Int64},
		{StringValue: "true", Kind: reflect.Int64},
		{StringValue: "test", Kind: reflect.Uint},
		{StringValue: "-6", Kind: reflect.Uint},
		{StringValue: "test", Kind: reflect.Uint8},
		{StringValue: "-6", Kind: reflect.Uint8},
		{StringValue: "test", Kind: reflect.Uint16},
		{StringValue: "-6", Kind: reflect.Uint16},
		{StringValue: "test", Kind: reflect.Uint32},
		{StringValue: "-6", Kind: reflect.Uint32},
		{StringValue: "test", Kind: reflect.Uint64},
		{StringValue: "-6", Kind: reflect.Uint64},
	}

	for _, test := range tests {
		parsedValue, err := getParsedValue(test.Kind, test.StringValue)

		if err == nil {
			t.Errorf("unexpected unhandled error - stringValue: %v, Kind: %v", test.StringValue, test.Kind)
			continue
		}

		if parsedValue != reflect.ValueOf(nil) {
			t.Errorf("unexpected parsed value - parsedValue: %v", parsedValue)
		}
	}
}
