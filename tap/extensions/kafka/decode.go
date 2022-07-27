package kafka

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
)

type discarder interface {
	Discard(int) (int, error)
}

type decoder struct {
	reader io.Reader
	remain int
	buffer [8]byte
	err    error
	table  *crc32.Table
	crc32  uint32
}

func (d *decoder) Read(b []byte) (int, error) {
	if d.err != nil {
		return 0, d.err
	}
	if d.remain == 0 {
		return 0, io.EOF
	}
	if len(b) > d.remain {
		b = b[:d.remain]
	}
	n, err := d.reader.Read(b)
	if n > 0 && d.table != nil {
		d.crc32 = crc32.Update(d.crc32, d.table, b[:n])
	}
	d.remain -= n
	return n, err
}

func (d *decoder) ReadByte() (byte, error) {
	c := d.readByte()
	return c, d.err
}

func (d *decoder) decodeBool(v value) {
	v.setBool(d.readBool())
}

func (d *decoder) decodeInt8(v value) {
	v.setInt8(d.readInt8())
}

func (d *decoder) decodeInt16(v value) {
	v.setInt16(d.readInt16())
}

func (d *decoder) decodeInt32(v value) {
	v.setInt32(d.readInt32())
}

func (d *decoder) decodeInt64(v value) {
	v.setInt64(d.readInt64())
}

func (d *decoder) decodeString(v value) {
	v.setString(d.readString())
}

func (d *decoder) decodeCompactString(v value) {
	v.setString(d.readCompactString())
}

func (d *decoder) decodeBytes(v value) {
	v.setBytes(d.readBytes())
}

func (d *decoder) decodeCompactBytes(v value) {
	v.setBytes(d.readCompactBytes())
}

func (d *decoder) decodeArray(v value, elemType reflect.Type, decodeElem decodeFunc) {
	if n := d.readInt32(); n < 0 || n > 65535 {
		v.setArray(array{})
	} else {
		a := makeArray(elemType, int(n))
		for i := 0; i < int(n) && d.remain > 0; i++ {
			decodeElem(d, a.index(i))
		}
		v.setArray(a)
	}
}

func (d *decoder) decodeCompactArray(v value, elemType reflect.Type, decodeElem decodeFunc) {
	if n := d.readUnsignedVarInt(); n < 1 || n > 65535 {
		v.setArray(array{})
	} else {
		a := makeArray(elemType, int(n-1))
		for i := 0; i < int(n-1) && d.remain > 0; i++ {
			decodeElem(d, a.index(i))
		}
		v.setArray(a)
	}
}

func (d *decoder) decodeRecordV0(v value) {
	x := &RecordV0{}
	x.Unknown = d.readInt8()
	x.Attributes = d.readInt8()
	x.TimestampDelta = d.readInt8()
	x.OffsetDelta = d.readInt8()

	x.KeyLength = int8(d.readVarInt())
	key := strings.Builder{}
	for i := 0; i < int(x.KeyLength); i++ {
		key.WriteString(fmt.Sprintf("%c", d.readInt8()))
	}
	x.Key = key.String()

	x.ValueLen = int8(d.readVarInt())
	value := strings.Builder{}
	for i := 0; i < int(x.ValueLen); i++ {
		value.WriteString(fmt.Sprintf("%c", d.readInt8()))
	}
	x.Value = value.String()

	headerLen := d.readInt8() / 2
	headers := make([]RecordHeader, 0)
	for i := 0; i < int(headerLen); i++ {
		header := &RecordHeader{}

		header.HeaderKeyLength = int8(d.readVarInt())
		headerKey := strings.Builder{}
		for j := 0; j < int(header.HeaderKeyLength); j++ {
			headerKey.WriteString(fmt.Sprintf("%c", d.readInt8()))
		}
		header.HeaderKey = headerKey.String()

		header.HeaderValueLength = int8(d.readVarInt())
		headerValue := strings.Builder{}
		for j := 0; j < int(header.HeaderValueLength); j++ {
			headerValue.WriteString(fmt.Sprintf("%c", d.readInt8()))
		}
		header.Value = headerValue.String()

		headers = append(headers, *header)
	}
	x.Headers = headers

	v.val.Set(valueOf(x).val)
}

func (d *decoder) discardAll() {
	d.discard(d.remain)
}

func (d *decoder) discard(n int) {
	if n > d.remain {
		n = d.remain
	}
	var err error
	if r, _ := d.reader.(discarder); r != nil {
		n, err = r.Discard(n)
		d.remain -= n
	} else {
		_, err = io.Copy(ioutil.Discard, d)
	}
	d.setError(err)
}

func (d *decoder) read(n int) []byte {
	b := make([]byte, n)
	n, err := io.ReadFull(d, b)
	b = b[:n]
	d.setError(err)
	return b
}

func (d *decoder) setError(err error) {
	if d.err == nil && err != nil {
		d.err = err
		d.discardAll()
	}
}

func (d *decoder) readFull(b []byte) bool {
	n, err := io.ReadFull(d, b)
	d.setError(err)
	return n == len(b)
}

func (d *decoder) readByte() byte {
	if d.readFull(d.buffer[:1]) {
		return d.buffer[0]
	}
	return 0
}

func (d *decoder) readBool() bool {
	return d.readByte() != 0
}

func (d *decoder) readInt8() int8 {
	if d.readFull(d.buffer[:1]) {
		return decodeReadInt8(d.buffer[:1])
	}
	return 0
}

func (d *decoder) readInt16() int16 {
	if d.readFull(d.buffer[:2]) {
		return decodeReadInt16(d.buffer[:2])
	}
	return 0
}

func (d *decoder) readInt32() int32 {
	if d.readFull(d.buffer[:4]) {
		return decodeReadInt32(d.buffer[:4])
	}
	return 0
}

func (d *decoder) readInt64() int64 {
	if d.readFull(d.buffer[:8]) {
		return decodeReadInt64(d.buffer[:8])
	}
	return 0
}

func (d *decoder) readString() string {
	if n := d.readInt16(); n < 0 {
		return ""
	} else {
		return bytesToString(d.read(int(n)))
	}
}

func (d *decoder) readCompactString() string {
	if n := d.readUnsignedVarInt(); n < 1 {
		return ""
	} else {
		return bytesToString(d.read(int(n - 1)))
	}
}

func (d *decoder) readBytes() []byte {
	if n := d.readInt32(); n < 0 {
		return nil
	} else {
		return d.read(int(n))
	}
}

func (d *decoder) readCompactBytes() []byte {
	if n := d.readUnsignedVarInt(); n < 1 {
		return nil
	} else {
		return d.read(int(n - 1))
	}
}

func (d *decoder) readVarInt() int64 {
	n := 11 // varints are at most 11 bytes

	if n > d.remain {
		n = d.remain
	}

	x := uint64(0)
	s := uint(0)

	for n > 0 {
		b := d.readByte()

		if (b & 0x80) == 0 {
			x |= uint64(b) << s
			return int64(x>>1) ^ -(int64(x) & 1)
		}

		x |= uint64(b&0x7f) << s
		s += 7
		n--
	}

	d.setError(fmt.Errorf("cannot decode varint from input stream"))
	return 0
}

func (d *decoder) readUnsignedVarInt() uint64 {
	n := 11 // varints are at most 11 bytes

	if n > d.remain {
		n = d.remain
	}

	x := uint64(0)
	s := uint(0)

	for n > 0 {
		b := d.readByte()

		if (b & 0x80) == 0 {
			x |= uint64(b) << s
			return x
		}

		x |= uint64(b&0x7f) << s
		s += 7
		n--
	}

	d.setError(fmt.Errorf("cannot decode unsigned varint from input stream"))
	return 0
}

type decodeFunc func(*decoder, value)

var (
	_ io.Reader     = (*decoder)(nil)
	_ io.ByteReader = (*decoder)(nil)

	readerFrom = reflect.TypeOf((*io.ReaderFrom)(nil)).Elem()
)

func decodeFuncOf(typ reflect.Type, version int16, flexible bool, tag structTag) decodeFunc {
	if reflect.PtrTo(typ).Implements(readerFrom) {
		return readerDecodeFuncOf(typ)
	}
	switch typ.Kind() {
	case reflect.Bool:
		return (*decoder).decodeBool
	case reflect.Int8:
		return (*decoder).decodeInt8
	case reflect.Int16:
		return (*decoder).decodeInt16
	case reflect.Int32:
		return (*decoder).decodeInt32
	case reflect.Int64:
		return (*decoder).decodeInt64
	case reflect.String:
		return stringDecodeFuncOf(flexible, tag)
	case reflect.Struct:
		return structDecodeFuncOf(typ, version, flexible)
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 { // []byte
			return bytesDecodeFuncOf(flexible, tag)
		}
		return arrayDecodeFuncOf(typ, version, flexible, tag)
	default:
		panic("unsupported type: " + typ.String())
	}
}

func stringDecodeFuncOf(flexible bool, tag structTag) decodeFunc {
	if flexible {
		// In flexible messages, all strings are compact
		return (*decoder).decodeCompactString
	}
	return (*decoder).decodeString
}

func bytesDecodeFuncOf(flexible bool, tag structTag) decodeFunc {
	if flexible {
		// In flexible messages, all arrays are compact
		return (*decoder).decodeCompactBytes
	}
	return (*decoder).decodeBytes
}

func structDecodeFuncOf(typ reflect.Type, version int16, flexible bool) decodeFunc {
	type field struct {
		decode decodeFunc
		index  index
		tagID  int
	}

	var fields []field
	taggedFields := map[int]*field{}

	if typ == reflect.TypeOf(RecordV0{}) {
		return (*decoder).decodeRecordV0
	}

	forEachStructField(typ, func(typ reflect.Type, index index, tag string) {
		forEachStructTag(tag, func(tag structTag) bool {
			if tag.MinVersion <= version && version <= tag.MaxVersion {
				f := field{
					decode: decodeFuncOf(typ, version, flexible, tag),
					index:  index,
					tagID:  tag.TagID,
				}

				if tag.TagID < -1 {
					// Normal required field
					fields = append(fields, f)
				} else {
					// Optional tagged field (flexible messages only)
					taggedFields[tag.TagID] = &f
				}
				return false
			}
			return true
		})
	})

	return func(d *decoder, v value) {
		for i := range fields {
			f := &fields[i]
			f.decode(d, v.fieldByIndex(f.index))
		}

		if flexible {
			// See https://cwiki.apache.org/confluence/display/KAFKA/KIP-482%3A+The+Kafka+Protocol+should+Support+Optional+Tagged+Fields
			// for details of tag buffers in "flexible" messages.
			n := int(d.readUnsignedVarInt())

			for i := 0; i < n; i++ {
				tagID := int(d.readUnsignedVarInt())
				size := int(d.readUnsignedVarInt())

				f, ok := taggedFields[tagID]
				if ok {
					f.decode(d, v.fieldByIndex(f.index))
				} else {
					d.read(size)
				}
			}
		}
	}
}

func arrayDecodeFuncOf(typ reflect.Type, version int16, flexible bool, tag structTag) decodeFunc {
	elemType := typ.Elem()
	elemFunc := decodeFuncOf(elemType, version, flexible, tag)
	if flexible {
		// In flexible messages, all arrays are compact
		return func(d *decoder, v value) { d.decodeCompactArray(v, elemType, elemFunc) }
	}

	return func(d *decoder, v value) { d.decodeArray(v, elemType, elemFunc) }
}

func readerDecodeFuncOf(typ reflect.Type) decodeFunc {
	typ = reflect.PtrTo(typ)
	return func(d *decoder, v value) {
		if d.err == nil {
			_, err := v.iface(typ).(io.ReaderFrom).ReadFrom(d)
			if err != nil {
				d.setError(err)
			}
		}
	}
}

func decodeReadInt8(b []byte) int8 {
	return int8(b[0])
}

func decodeReadInt16(b []byte) int16 {
	return int16(binary.BigEndian.Uint16(b))
}

func decodeReadInt32(b []byte) int32 {
	return int32(binary.BigEndian.Uint32(b))
}

func decodeReadInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}
