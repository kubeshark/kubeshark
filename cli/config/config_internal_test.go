package config

import (
	"fmt"
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

type FieldSetValues struct {
	SetValues  []string
	FieldName  string
	FieldValue interface{}
}

func TestMergeSetFlagNoSeparator(t *testing.T) {
	tests := []struct {
		Name      string
		SetValues []string
	}{
		{Name: "empty value", SetValues: []string{""}},
		{Name: "single char", SetValues: []string{"t"}},
		{Name: "combine empty value and single char", SetValues: []string{"", "t"}},
		{Name: "two values without separator", SetValues: []string{"test", "test:true"}},
		{Name: "four values without separator", SetValues: []string{"test", "test:true", "testing!", "true"}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			configMock := ConfigMock{}
			configMockElemValue := reflect.ValueOf(&configMock).Elem()

			err := mergeSetFlag(configMockElemValue, test.SetValues)

			if err == nil {
				t.Errorf("unexpected unhandled error - SetValues: %v", test.SetValues)
				return
			}

			for i := 0; i < configMockElemValue.NumField(); i++ {
				currentField := configMockElemValue.Type().Field(i)
				currentFieldByName := configMockElemValue.FieldByName(currentField.Name)

				if !currentFieldByName.IsZero() {
					t.Errorf("unexpected value with not default value - SetValues: %v", test.SetValues)
				}
			}
		})
	}
}

func TestMergeSetFlagInvalidFlagName(t *testing.T) {
	tests := []struct {
		Name      string
		SetValues []string
	}{
		{Name: "invalid flag name", SetValues: []string{"invalid_flag=true"}},
		{Name: "invalid flag name inside section struct", SetValues: []string{"section.invalid_flag=test"}},
		{Name: "flag name is a struct", SetValues: []string{"section=test"}},
		{Name: "empty flag name", SetValues: []string{"=true"}},
		{Name: "four tests combined", SetValues: []string{"invalid_flag=true", "config.invalid_flag=test", "section=test", "=true"}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			configMock := ConfigMock{}
			configMockElemValue := reflect.ValueOf(&configMock).Elem()

			err := mergeSetFlag(configMockElemValue, test.SetValues)

			if err == nil {
				t.Errorf("unexpected unhandled error - SetValues: %v", test.SetValues)
				return
			}

			for i := 0; i < configMockElemValue.NumField(); i++ {
				currentField := configMockElemValue.Type().Field(i)
				currentFieldByName := configMockElemValue.FieldByName(currentField.Name)

				if !currentFieldByName.IsZero() {
					t.Errorf("unexpected case - SetValues: %v", test.SetValues)
				}
			}
		})
	}
}

func TestMergeSetFlagInvalidFlagValue(t *testing.T) {
	tests := []struct {
		Name      string
		SetValues []string
	}{
		{Name: "bool value to int field", SetValues: []string{"int-field=true"}},
		{Name: "int value to bool field", SetValues: []string{"bool-field:5"}},
		{Name: "int value to uint field", SetValues: []string{"uint-field=-1"}},
		{Name: "bool value to int slice field", SetValues: []string{"int-slice-field=true"}},
		{Name: "int value to bool slice field", SetValues: []string{"bool-slice-field=5"}},
		{Name: "int value to uint slice field", SetValues: []string{"uint-slice-field=-1"}},
		{Name: "int slice value to int field", SetValues: []string{"int-field=6", "int-field=66"}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			configMock := ConfigMock{}
			configMockElemValue := reflect.ValueOf(&configMock).Elem()

			err := mergeSetFlag(configMockElemValue, test.SetValues)

			if err == nil {
				t.Errorf("unexpected unhandled error - SetValues: %v", test.SetValues)
				return
			}

			for i := 0; i < configMockElemValue.NumField(); i++ {
				currentField := configMockElemValue.Type().Field(i)
				currentFieldByName := configMockElemValue.FieldByName(currentField.Name)

				if !currentFieldByName.IsZero() {
					t.Errorf("unexpected case - SetValues: %v", test.SetValues)
				}
			}
		})
	}
}

func TestMergeSetFlagNotSliceValues(t *testing.T) {
	tests := []struct {
		Name            string
		FieldsSetValues []FieldSetValues
	}{
		{Name: "string field", FieldsSetValues: []FieldSetValues{{SetValues: []string{"string-field=test"}, FieldName: "StringField", FieldValue: "test"}}},
		{Name: "int field", FieldsSetValues: []FieldSetValues{{SetValues: []string{"int-field=6"}, FieldName: "IntField", FieldValue: 6}}},
		{Name: "bool field", FieldsSetValues: []FieldSetValues{{SetValues: []string{"bool-field=true"}, FieldName: "BoolField", FieldValue: true}}},
		{Name: "uint field", FieldsSetValues: []FieldSetValues{{SetValues: []string{"uint-field=6"}, FieldName: "UintField", FieldValue: uint(6)}}},
		{Name: "four fields combined", FieldsSetValues: []FieldSetValues {
			{SetValues: []string{"string-field=test"}, FieldName: "StringField", FieldValue: "test"},
			{SetValues: []string{"int-field=6"}, FieldName: "IntField", FieldValue: 6},
			{SetValues: []string{"bool-field=true"}, FieldName: "BoolField", FieldValue: true},
			{SetValues: []string{"uint-field=6"}, FieldName: "UintField", FieldValue: uint(6)},
		}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			configMock := ConfigMock{}
			configMockElemValue := reflect.ValueOf(&configMock).Elem()

			var setValues []string
			for _, fieldSetValues := range test.FieldsSetValues {
				setValues = append(setValues, fieldSetValues.SetValues...)
			}

			err := mergeSetFlag(configMockElemValue, setValues)

			if err != nil {
				t.Errorf("unexpected error result - err: %v", err)
				return
			}

			for _, fieldSetValues := range test.FieldsSetValues {
				fieldValue := configMockElemValue.FieldByName(fieldSetValues.FieldName).Interface()
				if fieldValue != fieldSetValues.FieldValue {
					t.Errorf("unexpected result - expected: %v, actual: %v", fieldSetValues.FieldValue, fieldValue)
				}
			}
		})
	}
}

func TestMergeSetFlagSliceValues(t *testing.T) {
	tests := []struct {
		Name            string
		FieldsSetValues []FieldSetValues
	}{
		{Name: "string slice field single value", FieldsSetValues: []FieldSetValues{{SetValues: []string{"string-slice-field=test"}, FieldName: "StringSliceField", FieldValue: []string{"test"}}}},
		{Name: "int slice field single value", FieldsSetValues: []FieldSetValues{{SetValues: []string{"int-slice-field=6"}, FieldName: "IntSliceField", FieldValue: []int{6}}}},
		{Name: "bool slice field single value", FieldsSetValues: []FieldSetValues{{SetValues: []string{"bool-slice-field=true"}, FieldName: "BoolSliceField", FieldValue: []bool{true}}}},
		{Name: "uint slice field single value", FieldsSetValues: []FieldSetValues{{SetValues: []string{"uint-slice-field=6"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6)}}}},
		{Name: "four single value fields combined", FieldsSetValues: []FieldSetValues{
			{SetValues: []string{"string-slice-field=test"}, FieldName: "StringSliceField", FieldValue: []string{"test"}},
			{SetValues: []string{"int-slice-field=6"}, FieldName: "IntSliceField", FieldValue: []int{6}},
			{SetValues: []string{"bool-slice-field=true"}, FieldName: "BoolSliceField", FieldValue: []bool{true}},
			{SetValues: []string{"uint-slice-field=6"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6)}},
		}},
		{Name: "string slice field two values", FieldsSetValues: []FieldSetValues{{SetValues: []string{"string-slice-field=test", "string-slice-field=test2"}, FieldName: "StringSliceField", FieldValue: []string{"test", "test2"}}}},
		{Name: "int slice field two values", FieldsSetValues: []FieldSetValues{{SetValues: []string{"int-slice-field=6", "int-slice-field=66"}, FieldName: "IntSliceField", FieldValue: []int{6, 66}}}},
		{Name: "bool slice field two values", FieldsSetValues: []FieldSetValues{{SetValues: []string{"bool-slice-field=true", "bool-slice-field=false"}, FieldName: "BoolSliceField", FieldValue: []bool{true, false}}}},
		{Name: "uint slice field two values", FieldsSetValues: []FieldSetValues{{SetValues: []string{"uint-slice-field=6", "uint-slice-field=66"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6), uint(66)}}}},
		{Name: "four two values fields combined", FieldsSetValues: []FieldSetValues{
			{SetValues: []string{"string-slice-field=test", "string-slice-field=test2"}, FieldName: "StringSliceField", FieldValue: []string{"test", "test2"}},
			{SetValues: []string{"int-slice-field=6", "int-slice-field=66"}, FieldName: "IntSliceField", FieldValue: []int{6, 66}},
			{SetValues: []string{"bool-slice-field=true", "bool-slice-field=false"}, FieldName: "BoolSliceField", FieldValue: []bool{true, false}},
			{SetValues: []string{"uint-slice-field=6", "uint-slice-field=66"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6), uint(66)}},
		}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			configMock := ConfigMock{}
			configMockElemValue := reflect.ValueOf(&configMock).Elem()

			var setValues []string
			for _, fieldSetValues := range test.FieldsSetValues {
				setValues = append(setValues, fieldSetValues.SetValues...)
			}

			err := mergeSetFlag(configMockElemValue, setValues)

			if err != nil {
				t.Errorf("unexpected error result - err: %v", err)
				return
			}

			for _, fieldSetValues := range test.FieldsSetValues {
				fieldValue := configMockElemValue.FieldByName(fieldSetValues.FieldName).Interface()
				if !reflect.DeepEqual(fieldValue, fieldSetValues.FieldValue) {
					t.Errorf("unexpected result - expected: %v, actual: %v", fieldSetValues.FieldValue, fieldValue)
				}
			}
		})
	}
}

func TestMergeSetFlagMixValues(t *testing.T) {
	tests := []struct {
		Name            string
		FieldsSetValues []FieldSetValues
	}{
		{Name: "single value all fields", FieldsSetValues: []FieldSetValues{
			{SetValues: []string{"string-slice-field=test"}, FieldName: "StringSliceField", FieldValue: []string{"test"}},
			{SetValues: []string{"int-slice-field=6"}, FieldName: "IntSliceField", FieldValue: []int{6}},
			{SetValues: []string{"bool-slice-field=true"}, FieldName: "BoolSliceField", FieldValue: []bool{true}},
			{SetValues: []string{"uint-slice-field=6"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6)}},
			{SetValues: []string{"string-field=test"}, FieldName: "StringField", FieldValue: "test"},
			{SetValues: []string{"int-field=6"}, FieldName: "IntField", FieldValue: 6},
			{SetValues: []string{"bool-field=true"}, FieldName: "BoolField", FieldValue: true},
			{SetValues: []string{"uint-field=6"}, FieldName: "UintField", FieldValue: uint(6)},
		}},
		{Name: "two values slice fields and single value fields", FieldsSetValues: []FieldSetValues{
			{SetValues: []string{"string-slice-field=test", "string-slice-field=test2"}, FieldName: "StringSliceField", FieldValue: []string{"test", "test2"}},
			{SetValues: []string{"int-slice-field=6", "int-slice-field=66"}, FieldName: "IntSliceField", FieldValue: []int{6, 66}},
			{SetValues: []string{"bool-slice-field=true", "bool-slice-field=false"}, FieldName: "BoolSliceField", FieldValue: []bool{true, false}},
			{SetValues: []string{"uint-slice-field=6", "uint-slice-field=66"}, FieldName: "UintSliceField", FieldValue: []uint{uint(6), uint(66)}},
			{SetValues: []string{"string-field=test"}, FieldName: "StringField", FieldValue: "test"},
			{SetValues: []string{"int-field=6"}, FieldName: "IntField", FieldValue: 6},
			{SetValues: []string{"bool-field=true"}, FieldName: "BoolField", FieldValue: true},
			{SetValues: []string{"uint-field=6"}, FieldName: "UintField", FieldValue: uint(6)},
		}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			configMock := ConfigMock{}
			configMockElemValue := reflect.ValueOf(&configMock).Elem()

			var setValues []string
			for _, fieldSetValues := range test.FieldsSetValues {
				setValues = append(setValues, fieldSetValues.SetValues...)
			}

			err := mergeSetFlag(configMockElemValue, setValues)

			if err != nil {
				t.Errorf("unexpected error result - err: %v", err)
				return
			}

			for _, fieldSetValues := range test.FieldsSetValues {
				fieldValue := configMockElemValue.FieldByName(fieldSetValues.FieldName).Interface()
				if !reflect.DeepEqual(fieldValue, fieldSetValues.FieldValue) {
					t.Errorf("unexpected result - expected: %v, actual: %v", fieldSetValues.FieldValue, fieldValue)
				}
			}
		})
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
		t.Run(fmt.Sprintf("%v %v", test.Kind, test.StringValue), func(t *testing.T) {
			parsedValue, err := getParsedValue(test.Kind, test.StringValue)

			if err != nil {
				t.Errorf("unexpected error result - err: %v", err)
				return
			}

			if parsedValue.Interface() != test.ActualValue {
				t.Errorf("unexpected result - expected: %v, actual: %v", test.ActualValue, parsedValue)
			}
		})
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
		t.Run(fmt.Sprintf("%v %v", test.Kind, test.StringValue), func(t *testing.T) {
			parsedValue, err := getParsedValue(test.Kind, test.StringValue)

			if err == nil {
				t.Errorf("unexpected unhandled error - stringValue: %v, Kind: %v", test.StringValue, test.Kind)
				return
			}

			if parsedValue != reflect.ValueOf(nil) {
				t.Errorf("unexpected parsed value - parsedValue: %v", parsedValue)
			}
		})
	}
}
