package goUtils

import (
	"log"
	"reflect"
	"runtime/debug"
)

func HandleExcWrapper(fn interface{}, params ...interface{}) (result []reflect.Value) {
	defer func() {
		if panicMessage := recover(); panicMessage != nil {
			stack := debug.Stack()
			log.Fatalf("Unhandled panic: %v\n stack: %s", panicMessage, stack)
		}
	}()
	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != len(params) {
		panic("incorrect number of parameters!")
	}
	inputs := make([]reflect.Value, len(params))
	for k, in := range params {
		inputs[k] = reflect.ValueOf(in)
	}
	return f.Call(inputs)
}
