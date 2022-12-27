package goUtils

import (
	"reflect"

	"github.com/rs/zerolog/log"
)

func HandleExcWrapper(fn interface{}, params ...interface{}) (result []reflect.Value) {
	defer func() {
		if panicMessage := recover(); panicMessage != nil {
			log.Fatal().
				Stack().
				Interface("msg", panicMessage).
				Msg("Unhandled panic!")
		}
	}()
	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != len(params) {
		panic("Incorrect number of parameters!")
	}
	inputs := make([]reflect.Value, len(params))
	for k, in := range params {
		inputs[k] = reflect.ValueOf(in)
	}
	return f.Call(inputs)
}
