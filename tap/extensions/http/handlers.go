package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/up9inc/mizu/tap/api"
)

func filterAndEmit(item *api.OutputChannelItem, emitter api.Emitter, options *api.TrafficFilteringOptions) {
	if IsIgnoredUserAgent(item, options) {
		return
	}

	if !options.DisableRedaction {
		FilterSensitiveData(item, options)
	}

	replaceForwardedFor(item)

	emitter.Emit(item)
}

func replaceForwardedFor(item *api.OutputChannelItem) {
	if item.Protocol.Name != "http" {
		return
	}

	request := item.Pair.Request.Payload.(api.HTTPPayload).Data.(*http.Request)

	forwardedFor := request.Header.Get("X-Forwarded-For")
	if forwardedFor == "" {
		return
	}

	ips := strings.Split(forwardedFor, ",")
	lastIP := strings.TrimSpace(ips[len(ips) - 1])

	item.ConnectionInfo.ClientIP = lastIP
	item.ConnectionInfo.ClientPort = ""
}

func handleHTTP2Stream(http2Assembler *Http2Assembler, tcpID *api.TcpID, superTimer *api.SuperTimer, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	streamID, messageHTTP1, isGrpc, err := http2Assembler.readMessage()
	if err != nil {
		return err
	}

	var item *api.OutputChannelItem

	switch messageHTTP1 := messageHTTP1.(type) {
	case http.Request:
		ident := fmt.Sprintf(
			"%s->%s %s->%s %d %s",
			tcpID.SrcIP,
			tcpID.DstIP,
			tcpID.SrcPort,
			tcpID.DstPort,
			streamID,
			"HTTP2",
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
			"%s->%s %s->%s %d %s",
			tcpID.DstIP,
			tcpID.SrcIP,
			tcpID.DstPort,
			tcpID.SrcPort,
			streamID,
			"HTTP2",
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
		if isGrpc {
			item.Protocol = grpcProtocol
		} else {
			item.Protocol = http2Protocol
		}
		filterAndEmit(item, emitter, options)
	}

	return nil
}

func handleHTTP1ClientStream(b *bufio.Reader, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, emitter api.Emitter, options *api.TrafficFilteringOptions) (switchingProtocolsHTTP2 bool, req *http.Request, err error) {
	req, err = http.ReadRequest(b)
	if err != nil {
		return
	}
	counterPair.Request++

	// Check HTTP2 upgrade - HTTP2 Over Cleartext (H2C)
	if strings.Contains(strings.ToLower(req.Header.Get("Connection")), "upgrade") && strings.ToLower(req.Header.Get("Upgrade")) == "h2c" {
		switchingProtocolsHTTP2 = true
	}

	var body []byte
	body, err = ioutil.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind

	ident := fmt.Sprintf(
		"%s->%s %s->%s %d %s",
		tcpID.SrcIP,
		tcpID.DstIP,
		tcpID.SrcPort,
		tcpID.DstPort,
		counterPair.Request,
		"HTTP1",
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
	return
}

func handleHTTP1ServerStream(b *bufio.Reader, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, emitter api.Emitter, options *api.TrafficFilteringOptions) (switchingProtocolsHTTP2 bool, err error) {
	var res *http.Response
	res, err = http.ReadResponse(b, nil)
	if err != nil {
		return
	}
	counterPair.Response++

	// Check HTTP2 upgrade - HTTP2 Over Cleartext (H2C)
	if res.StatusCode == 101 && strings.Contains(strings.ToLower(res.Header.Get("Connection")), "upgrade") && strings.ToLower(res.Header.Get("Upgrade")) == "h2c" {
		switchingProtocolsHTTP2 = true
	}

	var body []byte
	body, err = ioutil.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind

	ident := fmt.Sprintf(
		"%s->%s %s->%s %d %s",
		tcpID.DstIP,
		tcpID.SrcIP,
		tcpID.DstPort,
		tcpID.SrcPort,
		counterPair.Response,
		"HTTP1",
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
	return
}
