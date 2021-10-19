package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/up9inc/mizu/tap/api"
)

func filterAndEmit(item *api.OutputChannelItem, emitter api.Emitter, options *api.TrafficFilteringOptions) {
	if IsIgnoredUserAgent(item, options) {
		return
	}

	if !options.DisableRedaction {
		FilterSensitiveData(item, options)
	}

	emitter.Emit(item)
}

func handleHTTP2Stream(grpcAssembler *GrpcAssembler, tcpID *api.TcpID, superTimer *api.SuperTimer, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	streamID, messageHTTP1, err := grpcAssembler.readMessage()
	if err != nil {
		return err
	}

	var item *api.OutputChannelItem

	switch messageHTTP1 := messageHTTP1.(type) {
	case http.Request:
		ident := fmt.Sprintf(
			"%s->%s %s->%s %d",
			tcpID.SrcIP,
			tcpID.DstIP,
			tcpID.SrcPort,
			tcpID.DstPort,
			streamID,
		)
		item = reqResMatcher.registerRequest(ident, &messageHTTP1, superTimer.CaptureTime)
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
		ident := fmt.Sprintf(
			"%s->%s %s->%s %d",
			tcpID.DstIP,
			tcpID.SrcIP,
			tcpID.DstPort,
			tcpID.SrcPort,
			streamID,
		)
		item = reqResMatcher.registerResponse(ident, &messageHTTP1, superTimer.CaptureTime)
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
		item.Protocol = http2Protocol
		filterAndEmit(item, emitter, options)
	}

	return nil
}

func handleHTTP1ClientStream(b *bufio.Reader, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	req, err := http.ReadRequest(b)
	if err != nil {
		// log.Println("Error reading stream:", err)
		return err
	}
	counterPair.Request++

	body, err := ioutil.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
	s := len(body)
	if err != nil {
		log.Printf("[HTTP-request-body] stream %s Got body err: %s", tcpID.Ident, err)
	}
	if err := req.Body.Close(); err != nil {
		log.Printf("[HTTP-request-body-close] stream %s Failed to close request body: %s", tcpID.Ident, err)
	}
	encoding := req.Header["Content-Encoding"]
	log.Printf("HTTP/1 Request: %s %s %s (Body:%d) -> %s", tcpID.Ident, req.Method, req.URL, s, encoding)

	ident := fmt.Sprintf(
		"%s->%s %s->%s %d",
		tcpID.SrcIP,
		tcpID.DstIP,
		tcpID.SrcPort,
		tcpID.DstPort,
		counterPair.Request,
	)
	item := reqResMatcher.registerRequest(ident, req, superTimer.CaptureTime)
	if item != nil {
		item.ConnectionInfo = &api.ConnectionInfo{
			ClientIP:   tcpID.SrcIP,
			ClientPort: tcpID.SrcPort,
			ServerIP:   tcpID.DstIP,
			ServerPort: tcpID.DstPort,
			IsOutgoing: true,
		}
		filterAndEmit(item, emitter, options)
	}
	return nil
}

func handleHTTP1ServerStream(b *bufio.Reader, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	res, err := http.ReadResponse(b, nil)
	if err != nil {
		// log.Println("Error reading stream:", err)
		return err
	}
	counterPair.Response++

	body, err := ioutil.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
	s := len(body)
	if err != nil {
		log.Printf("[HTTP-response-body] HTTP/%s: failed to get body(parsed len:%d): %s", tcpID.Ident, s, err)
	}
	if err := res.Body.Close(); err != nil {
		log.Printf("[HTTP-response-body-close] HTTP/%s: failed to close body(parsed len:%d): %s", tcpID.Ident, s, err)
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
	log.Printf("HTTP/1 Response: %s %s (%d%s%d%s) -> %s", tcpID.Ident, res.Status, res.ContentLength, sym, s, contentType, encoding)

	ident := fmt.Sprintf(
		"%s->%s %s->%s %d",
		tcpID.DstIP,
		tcpID.SrcIP,
		tcpID.DstPort,
		tcpID.SrcPort,
		counterPair.Response,
	)
	item := reqResMatcher.registerResponse(ident, res, superTimer.CaptureTime)
	if item != nil {
		item.ConnectionInfo = &api.ConnectionInfo{
			ClientIP:   tcpID.DstIP,
			ClientPort: tcpID.DstPort,
			ServerIP:   tcpID.SrcIP,
			ServerPort: tcpID.SrcPort,
			IsOutgoing: false,
		}
		filterAndEmit(item, emitter, options)
	}
	return nil
}
