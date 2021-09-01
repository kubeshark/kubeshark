package config

import (
	"errors"
	"reflect"
	"strconv"
)

func getParsedValue(kind reflect.Kind, value string) (reflect.Value, error) {
	switch kind {
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Bool:
		boolArgumentValue, err := strconv.ParseBool(value)
		if err != nil {
			break
		}

		return reflect.ValueOf(boolArgumentValue), nil
	case reflect.Int:
		intArgumentValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			break
		}

		return reflect.ValueOf(int(intArgumentValue)), nil
	case reflect.Int8:
		intArgumentValue, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			break
		}

		return reflect.ValueOf(int8(intArgumentValue)), nil
	case reflect.Int16:
		intArgumentValue, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			break
		}

		return reflect.ValueOf(int16(intArgumentValue)), nil
	case reflect.Int32:
		intArgumentValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			break
		}

		return reflect.ValueOf(int32(intArgumentValue)), nil
	case reflect.Int64:
		intArgumentValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			break
		}

		return reflect.ValueOf(intArgumentValue), nil
	case reflect.Uint:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			break
		}

		return reflect.ValueOf(uint(uintArgumentValue)), nil
	case reflect.Uint8:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			break
		}

		return reflect.ValueOf(uint8(uintArgumentValue)), nil
	case reflect.Uint16:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			break
		}

		return reflect.ValueOf(uint16(uintArgumentValue)), nil
	case reflect.Uint32:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			break
		}

		return reflect.ValueOf(uint32(uintArgumentValue)), nil
	case reflect.Uint64:
		uintArgumentValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			break
		}

		return reflect.ValueOf(uintArgumentValue), nil
	}

	return reflect.ValueOf(nil), errors.New("value to parse does not match type")
}
