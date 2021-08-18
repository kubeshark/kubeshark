package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

func handleHTTP2Stream(grpcAssembler *GrpcAssembler, tcpID *api.TcpID, Emit func(item *api.OutputChannelItem)) error {
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
	}

	if item != nil {
		Emit(item)
	}

	return nil
}

func handleHTTP1ClientStream(b *bufio.Reader, tcpID *api.TcpID, Emit func(item *api.OutputChannelItem)) error {
	requestCounter++
	req, err := http.ReadRequest(b)
	if err != nil {
		log.Println("Error reading stream:", err)
		return err
	} else {
		body, _ := ioutil.ReadAll(req.Body)
		req.Body.Close()
		log.Printf("Received request: %+v with body: %+v\n", req, body)
	}

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
		Emit(item)
	}
	return nil
}

func handleHTTP1ServerStream(b *bufio.Reader, tcpID *api.TcpID, Emit func(item *api.OutputChannelItem)) error {
	responseCounter++
	res, err := http.ReadResponse(b, nil)
	if err != nil {
		log.Println("Error reading stream:", err)
		return err
	} else {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		log.Printf("Received response: %+v with body: %+v\n", res, body)
	}
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
		Emit(item)
	}
	return nil
}
