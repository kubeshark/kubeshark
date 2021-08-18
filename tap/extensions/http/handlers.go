package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

func handleHTTP2Stream(grpcAssembler *GrpcAssembler, tcpID *api.TcpID, emitter api.Emitter) error {
	streamID, messageHTTP1, err := grpcAssembler.readMessage()
	if err != nil {
		return err
	}

	var item *api.OutputChannelItem

	switch messageHTTP1 := messageHTTP1.(type) {
	case http.Request:
		requestCounter++
		ident := fmt.Sprintf(
			"%s->%s %s->%s %d",
			tcpID.SrcIP,
			tcpID.DstIP,
			tcpID.SrcPort,
			tcpID.DstPort,
			streamID,
		)
		item = reqResMatcher.registerRequest(ident, &messageHTTP1, time.Now())
		if item != nil {
			item.ConnectionInfo = &api.ConnectionInfo{
				ClientIP:   tcpID.SrcIP,
				ClientPort: tcpID.SrcPort,
				ServerIP:   tcpID.DstIP,
				ServerPort: tcpID.DstPort,
				IsOutgoing: true,
			}
		}
	case http.Response:
		responseCounter++
		ident := fmt.Sprintf(
			"%s->%s %s->%s %d",
			tcpID.DstIP,
			tcpID.SrcIP,
			tcpID.DstPort,
			tcpID.SrcPort,
			streamID,
		)
		item = reqResMatcher.registerResponse(ident, &messageHTTP1, time.Now())
		if item != nil {
			item.ConnectionInfo = &api.ConnectionInfo{
				ClientIP:   tcpID.DstIP,
				ClientPort: tcpID.DstPort,
				ServerIP:   tcpID.SrcIP,
				ServerPort: tcpID.SrcPort,
				IsOutgoing: false,
			}
		}
	}

	if item != nil {
		emitter.Emit(item)
	}

	return nil
}

func handleHTTP1ClientStream(b *bufio.Reader, tcpID *api.TcpID, emitter api.Emitter) error {
	requestCounter++
	req, err := http.ReadRequest(b)
	if err != nil {
		log.Println("Error reading stream:", err)
		return err
	}

	body, err := ioutil.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
	s := len(body)
	if err != nil {
		SilentError("HTTP-request-body", "stream %s Got body err: %s", tcpID.Ident, err)
	}
	if err := req.Body.Close(); err != nil {
		SilentError("HTTP-request-body-close", "stream %s Failed to close request body: %s", tcpID.Ident, err)
	}
	encoding := req.Header["Content-Encoding"]
	Debug("HTTP/1 Request: %s %s %s (Body:%d) -> %s", tcpID.Ident, req.Method, req.URL, s, encoding)

	ident := fmt.Sprintf(
		"%s->%s %s->%s %d",
		tcpID.SrcIP,
		tcpID.DstIP,
		tcpID.SrcPort,
		tcpID.DstPort,
		requestCounter,
	)
	item := reqResMatcher.registerRequest(ident, req, time.Now())
	if item != nil {
		item.ConnectionInfo = &api.ConnectionInfo{
			ClientIP:   tcpID.SrcIP,
			ClientPort: tcpID.SrcPort,
			ServerIP:   tcpID.DstIP,
			ServerPort: tcpID.DstPort,
			IsOutgoing: true,
		}
		emitter.Emit(item)
	}
	return nil
}

func handleHTTP1ServerStream(b *bufio.Reader, tcpID *api.TcpID, emitter api.Emitter) error {
	responseCounter++
	res, err := http.ReadResponse(b, nil)
	if err != nil {
		log.Println("Error reading stream:", err)
		return err
	}
	var req string
	req = fmt.Sprintf("<no-request-seen>")

	body, err := ioutil.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
	s := len(body)
	if err != nil {
		SilentError("HTTP-response-body", "HTTP/%s: failed to get body(parsed len:%d): %s", tcpID.Ident, s, err)
	}
	if err := res.Body.Close(); err != nil {
		SilentError("HTTP-response-body-close", "HTTP/%s: failed to close body(parsed len:%d): %s", tcpID.Ident, s, err)
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
	Debug("HTTP/1 Response: %s %s URL:%s (%d%s%d%s) -> %s", tcpID.Ident, res.Status, req, res.ContentLength, sym, s, contentType, encoding)

	ident := fmt.Sprintf(
		"%s->%s %s->%s %d",
		tcpID.DstIP,
		tcpID.SrcIP,
		tcpID.DstPort,
		tcpID.SrcPort,
		responseCounter,
	)
	item := reqResMatcher.registerResponse(ident, res, time.Now())
	if item != nil {
		item.ConnectionInfo = &api.ConnectionInfo{
			ClientIP:   tcpID.DstIP,
			ClientPort: tcpID.DstPort,
			ServerIP:   tcpID.SrcIP,
			ServerPort: tcpID.SrcPort,
			IsOutgoing: false,
		}
		emitter.Emit(item)
	}
	return nil
}
