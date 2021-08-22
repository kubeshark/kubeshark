package main

import (
	"log"
	"sync"
)

var reqResMatcher = CreateResponseRequestMatcher() // global

type RequestResponsePair struct {
	Request  Request
	Response Response
}

// Key is {client_addr}:{client_port}->{dest_addr}:{dest_port}::{correlation_id}
type requestResponseMatcher struct {
	openMessagesMap sync.Map
}

func CreateResponseRequestMatcher() requestResponseMatcher {
	newMatcher := &requestResponseMatcher{openMessagesMap: sync.Map{}}
	return *newMatcher
}

func (matcher *requestResponseMatcher) registerRequest(key string, request *Request) *RequestResponsePair {
	if response, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		return matcher.preparePair(request, response.(*Response))
	}

	matcher.openMessagesMap.Store(key, &request)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(key string, response *Response) *RequestResponsePair {
	if request, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		return matcher.preparePair(request.(*Request), response)
	}

	matcher.openMessagesMap.Store(key, &response)
	return nil
}

func (matcher *requestResponseMatcher) preparePair(request *Request, response *Response) *RequestResponsePair {
	return &RequestResponsePair{
		Request:  *request,
		Response: *response,
	}
}

func (reqResPair *RequestResponsePair) print() {
	log.Printf("----------------\n")
	reqResPair.Request.print()
	reqResPair.Response.print()
}
