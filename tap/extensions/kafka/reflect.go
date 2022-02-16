// +build !unsafe

package kafka

import (
	"reflect"
)

type index []int

type value struct {
	val reflect.Value
}

func valueOf(x interface{}) value {
	return value{val: reflect.ValueOf(x).Elem()}
}

func (v value) iface(t reflect.Type) interface{} { return v.val.Addr().Interface() }

func (v value) array(t reflect.Type) array { return array{val: v.val} } //nolint

func (v value) setBool(b bool) { v.val.SetBool(b) }

func (v value) setInt8(i int8) { v.setInt64(int64(i)) }

func (v value) setInt16(i int16) { v.setInt64(int64(i)) }

func (v value) setInt32(i int32) { v.setInt64(int64(i)) }

func (v value) setInt64(i int64) { v.val.SetInt(i) }

func (v value) setString(s string) { v.val.SetString(s) }

func (v value) setBytes(b []byte) { v.val.SetBytes(b) }

func (v value) setArray(a array) {
	if a.val.IsValid() {
		v.val.Set(a.val)
	} else {
		v.val.Set(reflect.Zero(v.val.Type()))
	}
}

func (v value) fieldByIndex(i index) value {
	return value{val: v.val.FieldByIndex(i)}
}

type array struct {
	val reflect.Value
}

func makeArray(t reflect.Type, n int) array {
	return array{val: reflect.MakeSlice(reflect.SliceOf(t), n, n)}
}

func (a array) index(i int) value { return value{val: a.val.Index(i)} }

func indexOf(s reflect.StructField) index { return index(s.Index) }

func bytesToString(b []byte) string { return string(b) }
