package kafka

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type readable interface {
	readFrom(*bufio.Reader, int) (int, error)
}

var errShortRead = errors.New("not enough bytes available to load the response")

func peekRead(r *bufio.Reader, sz int, n int, f func([]byte)) (int, error) {
	if n > sz {
		return sz, errShortRead
	}
	b, err := r.Peek(n)
	if err != nil {
		return sz, err
	}
	f(b)
	return discardN(r, sz, n)
}

func readInt8(r *bufio.Reader, sz int, v *int8) (int, error) {
	return peekRead(r, sz, 1, func(b []byte) { *v = makeInt8(b) })
}

func readInt16(r *bufio.Reader, sz int, v *int16) (int, error) {
	return peekRead(r, sz, 2, func(b []byte) { *v = makeInt16(b) })
}

func readInt32(r *bufio.Reader, sz int, v *int32) (int, error) {
	return peekRead(r, sz, 4, func(b []byte) { *v = makeInt32(b) })
}

func readInt64(r *bufio.Reader, sz int, v *int64) (int, error) {
	return peekRead(r, sz, 8, func(b []byte) { *v = makeInt64(b) })
}

func readString(r *bufio.Reader, sz int, v *string) (int, error) {
	return readStringWith(r, sz, func(r *bufio.Reader, sz int, n int) (remain int, err error) {
		*v, remain, err = readNewString(r, sz, n)
		return
	})
}

func readStringWith(r *bufio.Reader, sz int, cb func(*bufio.Reader, int, int) (int, error)) (int, error) {
	var err error
	var len int16

	if sz, err = readInt16(r, sz, &len); err != nil {
		return sz, err
	}

	n := int(len)
	if n > sz {
		return sz, errShortRead
	}

	return cb(r, sz, n)
}

func readNewString(r *bufio.Reader, sz int, n int) (string, int, error) {
	b, sz, err := readNewBytes(r, sz, n)
	return string(b), sz, err
}

func readBytes(r *bufio.Reader, sz int, v *[]byte) (int, error) {
	return readBytesWith(r, sz, func(r *bufio.Reader, sz int, n int) (remain int, err error) {
		*v, remain, err = readNewBytes(r, sz, n)
		return
	})
}

func readBytesWith(r *bufio.Reader, sz int, cb func(*bufio.Reader, int, int) (int, error)) (int, error) {
	var err error
	var n int

	if sz, err = readArrayLen(r, sz, &n); err != nil {
		return sz, err
	}

	if n > sz {
		return sz, errShortRead
	}

	return cb(r, sz, n)
}

func readNewBytes(r *bufio.Reader, sz int, n int) ([]byte, int, error) {
	var err error
	var b []byte
	var shortRead bool

	if n > 0 {
		if sz < n {
			n = sz
			shortRead = true
		}

		b = make([]byte, n)
		n, err = io.ReadFull(r, b)
		b = b[:n]
		sz -= n

		if err == nil && shortRead {
			err = errShortRead
		}
	}

	return b, sz, err
}

func readArrayLen(r *bufio.Reader, sz int, n *int) (int, error) {
	var err error
	var len int32
	if sz, err = readInt32(r, sz, &len); err != nil {
		return sz, err
	}
	*n = int(len)
	return sz, nil
}

func ReadAll(r *bufio.Reader, sz int, ptrs ...interface{}) (int, error) {
	var err error

	for _, ptr := range ptrs {
		if sz, err = readPtr(r, sz, ptr); err != nil {
			break
		}
	}

	return sz, err
}

func readPtr(r *bufio.Reader, sz int, ptr interface{}) (int, error) {
	switch v := ptr.(type) {
	case *int8:
		return readInt8(r, sz, v)
	case *int16:
		return readInt16(r, sz, v)
	case *int32:
		return readInt32(r, sz, v)
	case *int64:
		return readInt64(r, sz, v)
	case *string:
		return readString(r, sz, v)
	case *[]byte:
		return readBytes(r, sz, v)
	case readable:
		return v.readFrom(r, sz)
	default:
		panic(fmt.Sprintf("unsupported type: %T", v))
	}
}
