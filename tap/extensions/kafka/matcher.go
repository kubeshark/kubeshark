package kafka

import (
	"sync"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

type RequestResponsePair struct {
	Request  Request
	Response Response
}

// Key is {client_addr}_{client_port}_{dest_addr}_{dest_port}_{correlation_id}
type requestResponseMatcher struct {
	openMessagesMap *sync.Map
	maxTry          int
}

func createResponseRequestMatcher() api.RequestResponseMatcher {
	return &requestResponseMatcher{openMessagesMap: &sync.Map{}, maxTry: 3000}
}

func (matcher *requestResponseMatcher) GetMap() *sync.Map {
	return matcher.openMessagesMap
}

func (matcher *requestResponseMatcher) SetMaxTry(value int) {
	matcher.maxTry = value
}

func (matcher *requestResponseMatcher) registerRequest(key string, request *Request) *RequestResponsePair {
	if response, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		// Check for a situation that only occurs when a Kafka broker is initiating
		switch v := response.(type) {
		case *Response:
			return matcher.preparePair(request, v)
		}
	}

	matcher.openMessagesMap.Store(key, request)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(key string, response *Response) *RequestResponsePair {
	try := 0
	for {
		try++
		if try > matcher.maxTry {
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
