package main

import (
	"sync"
	"time"
)

var reqResMatcher = CreateResponseRequestMatcher() // global
const maxTry int = 3000

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

	matcher.openMessagesMap.Store(key, request)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(key string, response *Response) *RequestResponsePair {
	try := 0
	for {
		try++
		if try > maxTry {
			return nil
		}
		if request, found := matcher.openMessagesMap.LoadAndDelete(key); found {
			return matcher.preparePair(request.(*Request), response)
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (matcher *requestResponseMatcher) preparePair(request *Request, response *Response) *RequestResponsePair {
	return &RequestResponsePair{
		Request:  *request,
		Response: *response,
	}
}
