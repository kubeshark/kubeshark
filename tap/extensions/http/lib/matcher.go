package lib

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var reqResMatcher = createResponseRequestMatcher() // global

type requestResponsePair struct {
	Request  httpMessage `json:"request"`
	Response httpMessage `json:"response"`
}

type httpMessage struct {
	isRequest   bool
	captureTime time.Time
	orig        interface{}
}

// Key is {client_addr}:{client_port}->{dest_addr}:{dest_port}
type requestResponseMatcher struct {
	openMessagesMap sync.Map
}

func createResponseRequestMatcher() requestResponseMatcher {
	newMatcher := &requestResponseMatcher{openMessagesMap: sync.Map{}}
	return *newMatcher
}

func (matcher *requestResponseMatcher) registerRequest(ident string, request *http.Request, captureTime time.Time) *requestResponsePair {
	split := splitIdent(ident)
	key := genKey(split)

	requestHTTPMessage := httpMessage{
		isRequest:   true,
		captureTime: captureTime,
		orig:        request,
	}

	if response, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		// Type assertion always succeeds because all of the map's values are of httpMessage type
		responseHTTPMessage := response.(*httpMessage)
		if responseHTTPMessage.isRequest {
			SilentError("Request-Duplicate", "Got duplicate request with same identifier")
			return nil
		}
		Trace("Matched open Response for %s", key)
		return matcher.preparePair(&requestHTTPMessage, responseHTTPMessage)
	}

	matcher.openMessagesMap.Store(key, &requestHTTPMessage)
	Trace("Registered open Request for %s", key)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(ident string, response *http.Response, captureTime time.Time) *requestResponsePair {
	split := splitIdent(ident)
	key := genKey(split)

	responseHTTPMessage := httpMessage{
		isRequest:   false,
		captureTime: captureTime,
		orig:        response,
	}

	if request, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		// Type assertion always succeeds because all of the map's values are of httpMessage type
		requestHTTPMessage := request.(*httpMessage)
		if !requestHTTPMessage.isRequest {
			SilentError("Response-Duplicate", "Got duplicate response with same identifier")
			return nil
		}
		Trace("Matched open Request for %s", key)
		return matcher.preparePair(requestHTTPMessage, &responseHTTPMessage)
	}

	matcher.openMessagesMap.Store(key, &responseHTTPMessage)
	Trace("Registered open Response for %s", key)
	return nil
}

func (matcher *requestResponseMatcher) preparePair(requestHTTPMessage *httpMessage, responseHTTPMessage *httpMessage) *requestResponsePair {
	return &requestResponsePair{
		Request:  *requestHTTPMessage,
		Response: *responseHTTPMessage,
	}
}

func splitIdent(ident string) []string {
	ident = strings.Replace(ident, "->", " ", -1)
	return strings.Split(ident, " ")
}

func genKey(split []string) string {
	key := fmt.Sprintf("%s:%s->%s:%s,%s", split[0], split[2], split[1], split[3], split[4])
	return key
}

func (matcher *requestResponseMatcher) deleteOlderThan(t time.Time) int {
	numDeleted := 0

	matcher.openMessagesMap.Range(func(key interface{}, value interface{}) bool {
		message, _ := value.(*httpMessage)
		if message.captureTime.Before(t) {
			matcher.openMessagesMap.Delete(key)
			numDeleted++
		}
		return true
	})

	return numDeleted
}
