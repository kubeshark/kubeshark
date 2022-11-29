package goUtils

import (
	"reflect"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

func HandleExcWrapper(fn interface{}, params ...interface{}) (result []reflect.Value) {
	defer func() {
		if panicMessage := recover(); panicMessage != nil {
			stack := debug.Stack()
			log.Fatal().
				Interface("msg", panicMessage).
				Interface("stack", stack).
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
