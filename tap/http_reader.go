package tap

import (
	"bufio"
	"bytes"
	"compress/gzip"
	b64 "encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type httpReaderDataMsg struct {
	bytes     []byte
	timestamp time.Time
}

type tcpID struct {
	srcIP string
	dstIP string
	srcPort string
	dstPort string
}

func (tid *tcpID) String() string {
	return fmt.Sprintf("%s->%s %s->%s", tid.srcIP, tid.dstIP, tid.srcPort, tid.dstPort)
}

/* httpReader gets reads from a channel of bytes of tcp payload, and parses it into HTTP/1 requests and responses.
 * The payload is written to the channel by a tcpStream object that is dedicated to one tcp connection.
 * An httpReader object is unidirectional: it parses either a client stream or a server stream.
 * Implements io.Reader interface (Read)
 */
type httpReader struct {
	ident         string
	tcpID         tcpID
	isClient      bool
	isHTTP2       bool
	msgQueue      chan httpReaderDataMsg // Channel of captured reassembled tcp payload
	data          []byte
	captureTime   time.Time
	hexdump       bool
	parent        *tcpStream
	grpcAssembler GrpcAssembler
	messageCount  uint
	harWriter     *HarWriter
}

func (h *httpReader) Read(p []byte) (int, error) {
	var msg httpReaderDataMsg
	ok := true
	for ok && len(h.data) == 0 {
		msg, ok = <-h.msgQueue
		h.data = msg.bytes
		h.captureTime = msg.timestamp
	}
	if !ok || len(h.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, h.data)
	h.data = h.data[l:]
	return l, nil
}

func (h *httpReader) run(wg *sync.WaitGroup) {
	defer wg.Done()
	b := bufio.NewReader(h)

	if isHTTP2, err := checkIsHTTP2Connection(b, h.isClient); err != nil {
		SilentError("HTTP/2-Prepare-Connection", "stream %s Failed to check if client is HTTP/2: %s (%v,%+v)", h.ident, err, err, err)
		// Do something?
	} else {
		h.isHTTP2 = isHTTP2
	}

	if h.isHTTP2 {
		err := prepareHTTP2Connection(b, h.isClient)
		if err != nil {
			SilentError("HTTP/2-Prepare-Connection-After-Check", "stream %s error: %s (%v,%+v)", h.ident, err, err, err)
		}
		h.grpcAssembler = createGrpcAssembler(b)
	}

	for true {
		if h.isHTTP2 {
			err := h.handleHTTP2Stream()
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				SilentError("HTTP/2", "stream %s error: %s (%v,%+v)", h.ident, err, err, err)
				continue
			}
		} else if h.isClient {
			err := h.handleHTTP1ClientStream(b)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				SilentError("HTTP-request", "stream %s Request error: %s (%v,%+v)", h.ident, err, err, err)
				continue
			}
		} else {
			err := h.handleHTTP1ServerStream(b)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				SilentError("HTTP-response", "stream %s Response error: %s (%v,%+v)", h.ident, err, err, err)
				continue
			}
		}
	}
}

func (h *httpReader) handleHTTP2Stream() error {
	streamID, messageHTTP1, body, err := h.grpcAssembler.readMessage()
	h.messageCount++
	if err != nil {
		return err
	}

	var reqResPair *envoyMessageWrapper

	switch messageHTTP1 := messageHTTP1.(type) {
	case http.Request:
		ident := fmt.Sprintf("%s->%s %s->%s %d", h.tcpID.srcIP, h.tcpID.dstIP, h.tcpID.srcPort, h.tcpID.dstPort, streamID)
		reqResPair = reqResMatcher.registerRequest(ident, &messageHTTP1, h.captureTime, body, true)
	case http.Response:
		ident := fmt.Sprintf("%s->%s %s->%s %d", h.tcpID.dstIP, h.tcpID.srcIP, h.tcpID.dstPort, h.tcpID.srcPort, streamID)
		reqResPair = reqResMatcher.registerResponse(ident, &messageHTTP1, h.captureTime, body, true)
	}

	if reqResPair != nil {
		statsTracker.incMatchedMessages()

		if h.harWriter != nil {
			h.harWriter.WritePair(
				reqResPair.HttpBufferedTrace.Request.orig.(*http.Request),
				reqResPair.HttpBufferedTrace.Request.captureTime,
				reqResPair.HttpBufferedTrace.Response.orig.(*http.Response),
				reqResPair.HttpBufferedTrace.Response.captureTime,
				&reqResPair.HttpBufferedTrace.Request.connection,
			)
		}
	}

	return nil
}

func (h *httpReader) handleHTTP1ClientStream(b *bufio.Reader) error {
	req, err := http.ReadRequest(b)
	h.messageCount++
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
	s := len(body)
	if err != nil {
		SilentError("HTTP-request-body", "stream %s Got body err: %s", h.ident, err)
	} else if h.hexdump {
		Info("Body(%d/0x%x) - %s", len(body), len(body), hex.Dump(body))
	}
	if err := req.Body.Close(); err != nil {
		SilentError("HTTP-request-body-close", "stream %s Failed to close request body: %s", h.ident, err)
	}
	encoding := req.Header["Content-Encoding"]
	bodyStr, err := readBody(body, encoding)
	if err != nil {
		SilentError("HTTP-request-body-decode", "stream %s Failed to decode body: %s", h.ident, err)
	}
	Info("HTTP/%s Request: %s %s (Body:%d)", h.ident, req.Method, req.URL, s)

	ident := fmt.Sprintf("%s->%s %s->%s %d", h.tcpID.srcIP, h.tcpID.dstIP, h.tcpID.srcPort, h.tcpID.dstPort, h.messageCount)
	reqResPair := reqResMatcher.registerRequest(ident, req, h.captureTime, bodyStr, false)
	if reqResPair != nil {
		statsTracker.incMatchedMessages()

		if h.harWriter != nil {
			h.harWriter.WritePair(
				reqResPair.HttpBufferedTrace.Request.orig.(*http.Request),
				reqResPair.HttpBufferedTrace.Request.captureTime,
				reqResPair.HttpBufferedTrace.Response.orig.(*http.Response),
				reqResPair.HttpBufferedTrace.Response.captureTime,
				&reqResPair.HttpBufferedTrace.Request.connection,
			)
		}
	}

	h.parent.Lock()
	h.parent.urls = append(h.parent.urls, req.URL.String())
	h.parent.Unlock()

	return nil
}

func (h *httpReader) handleHTTP1ServerStream(b *bufio.Reader) error {
	res, err := http.ReadResponse(b, nil)
	h.messageCount++
	var req string
	h.parent.Lock()
	if len(h.parent.urls) == 0 {
		req = fmt.Sprintf("<no-request-seen>")
	} else {
		req, h.parent.urls = h.parent.urls[0], h.parent.urls[1:]
	}
	h.parent.Unlock()
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
	s := len(body)
	if err != nil {
		SilentError("HTTP-response-body", "HTTP/%s: failed to get body(parsed len:%d): %s", h.ident, s, err)
	}
	if h.hexdump {
		Info("Body(%d/0x%x) - %s", len(body), len(body), hex.Dump(body))
	}
	if err := res.Body.Close(); err != nil {
		SilentError("HTTP-response-body-close", "HTTP/%s: failed to close body(parsed len:%d): %s", h.ident, s, err)
	}
	sym := ","
	if res.ContentLength > 0 && res.ContentLength != int64(s) {
		sym = "!="
	}
	contentType, ok := res.Header["Content-Type"]
	if !ok {
		contentType = []string{http.DetectContentType(body)}
	}
	encoding := res.Header["Content-Encoding"]
	Info("HTTP/%s Response: %s URL:%s (%d%s%d%s) -> %s", h.ident, res.Status, req, res.ContentLength, sym, s, contentType, encoding)
	bodyStr, err := readBody(body, encoding)
	if err != nil {
		SilentError("HTTP-response-body-decode", "stream %s Failed to decode body: %s", h.ident, err)
	}

	ident := fmt.Sprintf("%s->%s %s->%s %d", h.tcpID.dstIP, h.tcpID.srcIP, h.tcpID.dstPort, h.tcpID.srcPort, h.messageCount)
	reqResPair := reqResMatcher.registerResponse(ident, res, h.captureTime, bodyStr, false)
	if reqResPair != nil {
		statsTracker.incMatchedMessages()

		if h.harWriter != nil {
			h.harWriter.WritePair(
				reqResPair.HttpBufferedTrace.Request.orig.(*http.Request),
				reqResPair.HttpBufferedTrace.Request.captureTime,
				reqResPair.HttpBufferedTrace.Response.orig.(*http.Response),
				reqResPair.HttpBufferedTrace.Response.captureTime,
				&reqResPair.HttpBufferedTrace.Request.connection,
			)
		}
	}

	return nil
}

func readBody(bodyBytes []byte, encoding []string) (string, error) {
	var bodyBuffer io.Reader
	bodyBuffer = bytes.NewBuffer(bodyBytes)
	var err error
	if len(encoding) > 0 && (encoding[0] == "gzip" || encoding[0] == "deflate") {
		bodyBuffer, err = gzip.NewReader(bodyBuffer)
		if err != nil {
			SilentError("HTTP-gunzip", "Failed to gzip decode: %s", err)
			return "", err
		}
	}
	if _, ok := bodyBuffer.(*gzip.Reader); ok {
		err = bodyBuffer.(*gzip.Reader).Close()
		if err != nil {
			return "", err
		}
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(bodyBuffer)
	return b64.StdEncoding.EncodeToString(buf.Bytes()), err
}
