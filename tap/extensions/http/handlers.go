package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

func handleHTTP2Stream(grpcAssembler *GrpcAssembler, tcpID *api.TcpID) (*api.RequestResponsePair, error) {
	streamID, messageHTTP1, err := grpcAssembler.readMessage()
	if err != nil {
		return nil, err
	}

	var reqResPair *api.RequestResponsePair

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
		reqResPair = reqResMatcher.registerRequest(ident, &messageHTTP1, time.Now())
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
		reqResPair = reqResMatcher.registerResponse(ident, &messageHTTP1, time.Now())
	}

	if reqResPair != nil {
		return reqResPair, nil
	}

	return nil, nil
}

func handleHTTP1ClientStream(b *bufio.Reader, tcpID *api.TcpID) error {
	requestCounter++
	req, err := http.ReadRequest(b)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil
	} else if err != nil {
		log.Println("Error reading stream:", err)
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
	reqResMatcher.registerRequest(ident, req, time.Now())
	return err
}

func handleHTTP1ServerStream(b *bufio.Reader, tcpID *api.TcpID) (*api.RequestResponsePair, error) {
	responseCounter++
	res, err := http.ReadResponse(b, nil)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, nil
	} else if err != nil {
		log.Println("Error reading stream:", err)
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
	reqResPair := reqResMatcher.registerResponse(ident, res, time.Now())
	if reqResPair != nil {
		return reqResPair, nil
	}
	return nil, err
}
