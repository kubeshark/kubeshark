package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/romana/rlog"

	"github.com/up9inc/mizu/tap/api"
)

var reqResMatcher = createResponseRequestMatcher() // global

// Key is {client_addr}:{client_port}->{dest_addr}:{dest_port}_{incremental_counter}
type requestResponseMatcher struct {
	openMessagesMap sync.Map
}

func createResponseRequestMatcher() requestResponseMatcher {
	newMatcher := &requestResponseMatcher{openMessagesMap: sync.Map{}}
	return *newMatcher
}

func (matcher *requestResponseMatcher) registerRequest(ident string, request *http.Request, captureTime time.Time) *api.OutputChannelItem {
	split := splitIdent(ident)
	key := genKey(split)

	requestHTTPMessage := api.GenericMessage{
		IsRequest:   true,
		CaptureTime: captureTime,
		Payload: HTTPPayload{
			Type: "http_request",
			Data: request,
		},
	}

	if response, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		responseHTTPMessage := response.(*api.GenericMessage)
		if responseHTTPMessage.IsRequest {
			rlog.Debugf("[Request-Duplicate] Got duplicate request with same identifier")
			return nil
		}
		rlog.Tracef(1, "Matched open Response for %s", key)
		return matcher.preparePair(&requestHTTPMessage, responseHTTPMessage)
	}

	matcher.openMessagesMap.Store(key, &requestHTTPMessage)
	rlog.Tracef(1, "Registered open Request for %s", key)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(ident string, response *http.Response, captureTime time.Time) *api.OutputChannelItem {
	split := splitIdent(ident)
	key := genKey(split)

	responseHTTPMessage := api.GenericMessage{
		IsRequest:   false,
		CaptureTime: captureTime,
		Payload: HTTPPayload{
			Type: "http_response",
			Data: response,
		},
	}

	if request, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		requestHTTPMessage := request.(*api.GenericMessage)
		if !requestHTTPMessage.IsRequest {
			rlog.Debugf("[Response-Duplicate] Got duplicate response with same identifier")
			return nil
		}
		rlog.Tracef(1, "Matched open Request for %s", key)
		return matcher.preparePair(requestHTTPMessage, &responseHTTPMessage)
	}

	matcher.openMessagesMap.Store(key, &responseHTTPMessage)
	rlog.Tracef(1, "Registered open Response for %s", key)
	return nil
}

func (matcher *requestResponseMatcher) preparePair(requestHTTPMessage *api.GenericMessage, responseHTTPMessage *api.GenericMessage) *api.OutputChannelItem {
	return &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      time.Now().UnixNano() / int64(time.Millisecond),
		ConnectionInfo: nil,
		Pair: &api.RequestResponsePair{
			Request:  *requestHTTPMessage,
			Response: *responseHTTPMessage,
		},
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
		message, _ := value.(*api.GenericMessage)
		if message.CaptureTime.Before(t) {
			matcher.openMessagesMap.Delete(key)
			numDeleted++
		}
		return true
	})

	return numDeleted
}
