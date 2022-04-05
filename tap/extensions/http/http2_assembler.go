package http

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

const frameHeaderLen = 9

var clientPreface = []byte(http2.ClientPreface)

const initialHeaderTableSize = 4096
const protoHTTP2 = "HTTP/2.0"
const protoMajorHTTP2 = 2
const protoMinorHTTP2 = 0

var maxHTTP2DataLen = 1 * 1024 * 1024 // 1MB

type messageFragment struct {
	headers []hpack.HeaderField
	data    []byte
}

type fragmentsByStream map[uint32]*messageFragment

func (fbs *fragmentsByStream) appendFrame(streamID uint32, frame http2.Frame) {
	switch frame := frame.(type) {
	case *http2.MetaHeadersFrame:
		if existingFragment, ok := (*fbs)[streamID]; ok {
			existingFragment.headers = append(existingFragment.headers, frame.Fields...)
		} else {
			// new fragment
			(*fbs)[streamID] = &messageFragment{headers: frame.Fields}
		}
	case *http2.DataFrame:
		newDataLen := len(frame.Data())
		if existingFragment, ok := (*fbs)[streamID]; ok {
			existingDataLen := len(existingFragment.data)
			// Never save more than maxHTTP2DataLen bytes
			numBytesToAppend := int(math.Min(float64(maxHTTP2DataLen-existingDataLen), float64(newDataLen)))

			existingFragment.data = append(existingFragment.data, frame.Data()[:numBytesToAppend]...)
		} else {
			// new fragment
			// In principle, should not happen with DATA frames, because they are always preceded by HEADERS

			// Never save more than maxHTTP2DataLen bytes
			numBytesToAppend := int(math.Min(float64(maxHTTP2DataLen), float64(newDataLen)))

			(*fbs)[streamID] = &messageFragment{data: frame.Data()[:numBytesToAppend]}
		}
	}
}

func (fbs *fragmentsByStream) pop(streamID uint32) ([]hpack.HeaderField, []byte) {
	headers := (*fbs)[streamID].headers
	data := (*fbs)[streamID].data
	delete(*fbs, streamID)

	return headers, data
}

func createHTTP2Assembler(b *bufio.Reader) *Http2Assembler {
	var framerOutput bytes.Buffer
	framer := http2.NewFramer(&framerOutput, b)
	framer.ReadMetaHeaders = hpack.NewDecoder(initialHeaderTableSize, nil)
	return &Http2Assembler{
		fragmentsByStream: make(fragmentsByStream),
		framer:            framer,
	}
}

type Http2Assembler struct {
	fragmentsByStream fragmentsByStream
	framer            *http2.Framer
}

func (ga *Http2Assembler) readMessage() (streamID uint32, messageHTTP1 interface{}, isGrpc bool, err error) {
	// Exactly one Framer is used for each half connection.
	// (Instead of creating a new Framer for each ReadFrame operation)
	// This is needed in order to decompress the headers,
	// because the compression context is updated with each requests/response.
	frame, err := ga.framer.ReadFrame()
	if err != nil {
		return
	}

	streamID = frame.Header().StreamID

	ga.fragmentsByStream.appendFrame(streamID, frame)

	if !(ga.isStreamEnd(frame)) {
		streamID = 0
		return
	}

	headers, data := ga.fragmentsByStream.pop(streamID)

	// Note: header keys are converted by http.Header.Set to canonical names, e.g. content-type -> Content-Type.
	// By converting the keys we violate the HTTP/2 specification, which state that all headers must be lowercase.
	headersHTTP1 := make(http.Header)
	for _, header := range headers {
		headersHTTP1.Add(header.Name, header.Value)
	}
	dataString := base64.StdEncoding.EncodeToString(data)

	// Use http1 types only because they are expected in http_matcher.
	method := headersHTTP1.Get(":method")
	status := headersHTTP1.Get(":status")

	// gRPC detection
	grpcStatus := headersHTTP1.Get("Grpc-Status")
	if grpcStatus != "" || strings.Contains(headersHTTP1.Get("Content-Type"), "application/grpc") {
		isGrpc = true
	}

	if method != "" {
		messageHTTP1 = http.Request{
			URL:           &url.URL{},
			Method:        method,
			Header:        headersHTTP1,
			Proto:         protoHTTP2,
			ProtoMajor:    protoMajorHTTP2,
			ProtoMinor:    protoMinorHTTP2,
			Body:          io.NopCloser(strings.NewReader(dataString)),
			ContentLength: int64(len(dataString)),
		}
	} else if status != "" {
		var statusCode int

		statusCode, err = strconv.Atoi(status)
		if err != nil {
			return
		}

		messageHTTP1 = http.Response{
			StatusCode:    statusCode,
			Header:        headersHTTP1,
			Proto:         protoHTTP2,
			ProtoMajor:    protoMajorHTTP2,
			ProtoMinor:    protoMinorHTTP2,
			Body:          io.NopCloser(strings.NewReader(dataString)),
			ContentLength: int64(len(dataString)),
		}
	} else {
		err = errors.New("failed to assemble stream: neither a request nor a message")
		return
	}

	return
}

func (ga *Http2Assembler) isStreamEnd(frame http2.Frame) bool {
	switch frame := frame.(type) {
	case *http2.MetaHeadersFrame:
		if frame.StreamEnded() {
			return true
		}
	case *http2.DataFrame:
		if frame.StreamEnded() {
			return true
		}
	}

	return false
}

/* Check if HTTP/2. Remove HTTP/2 client preface from start of buffer if present
 */
func checkIsHTTP2Connection(b *bufio.Reader, isClient bool) (bool, error) {
	if isClient {
		return checkIsHTTP2ClientStream(b)
	}

	return checkIsHTTP2ServerStream(b)
}

func prepareHTTP2Connection(b *bufio.Reader, isClient bool) error {
	if !isClient {
		return nil
	}

	return discardClientPreface(b)
}

func checkIsHTTP2ClientStream(b *bufio.Reader) (bool, error) {
	return checkClientPreface(b)
}

func checkIsHTTP2ServerStream(b *bufio.Reader) (bool, error) {
	buf, err := b.Peek(frameHeaderLen)
	if err != nil {
		return false, err
	}

	// If response starts with HTTP/1. then it's not HTTP/2
	if bytes.HasPrefix(buf, []byte("HTTP/1.")) {
		return false, nil
	}

	// Check server connection preface (a settings frame)
	frameHeader := http2.FrameHeader{
		Length:   uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2]),
		Type:     http2.FrameType(buf[3]),
		Flags:    http2.Flags(buf[4]),
		StreamID: binary.BigEndian.Uint32(buf[5:]) & (1<<31 - 1),
	}

	if frameHeader.Type != http2.FrameSettings {
		// If HTTP/2, but not start of stream, will also fulfill this condition.
		return false, nil
	}

	return true, nil
}

func checkClientPreface(b *bufio.Reader) (bool, error) {
	bytesStart, err := b.Peek(len(clientPreface))
	if err != nil {
		return false, err
	} else if len(bytesStart) != len(clientPreface) {
		return false, errors.New("checkClientPreface: not enough bytes read")
	}

	if !bytes.Equal(bytesStart, clientPreface) {
		return false, nil
	}

	return true, nil
}

func discardClientPreface(b *bufio.Reader) error {
	if isClientPrefacePresent, err := checkClientPreface(b); err != nil {
		return err
	} else if !isClientPrefacePresent {
		return errors.New("discardClientPreface: does not begin with client preface")
	}

	// Remove client preface string from the buffer
	n, err := b.Discard(len(clientPreface))
	if err != nil {
		return err
	} else if n != len(clientPreface) {
		return errors.New("discardClientPreface: failed to discard client preface")
	}

	return nil
}
