package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

// Key is {client_addr}_{client_port}_{dest_addr}_{dest_port}_{incremental_counter}_{proto_ident}
type requestResponseMatcher struct {
	openMessagesMap *sync.Map
}

func createResponseRequestMatcher() api.RequestResponseMatcher {
	return &requestResponseMatcher{openMessagesMap: &sync.Map{}}
}

func (matcher *requestResponseMatcher) GetMap() *sync.Map {
	return matcher.openMessagesMap
}

func (matcher *requestResponseMatcher) SetMaxTry(value int) {
}

func (matcher *requestResponseMatcher) registerRequest(ident string, request *http.Request, captureTime time.Time, captureSize int, protoMinor int) *api.OutputChannelItem {
	requestHTTPMessage := api.GenericMessage{
		IsRequest:   true,
		CaptureTime: captureTime,
		CaptureSize: captureSize,
		Payload: api.HTTPPayload{
			Type: TypeHttpRequest,
			Data: request,
		},
	}

	if response, found := matcher.openMessagesMap.LoadAndDelete(ident); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		responseHTTPMessage := response.(*api.GenericMessage)
		if responseHTTPMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(&requestHTTPMessage, responseHTTPMessage, protoMinor)
	}

	matcher.openMessagesMap.Store(ident, &requestHTTPMessage)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(ident string, response *http.Response, captureTime time.Time, captureSize int, protoMinor int) *api.OutputChannelItem {
	responseHTTPMessage := api.GenericMessage{
		IsRequest:   false,
		CaptureTime: captureTime,
		CaptureSize: captureSize,
		Payload: api.HTTPPayload{
			Type: TypeHttpResponse,
			Data: response,
		},
	}

	if request, found := matcher.openMessagesMap.LoadAndDelete(ident); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		requestHTTPMessage := request.(*api.GenericMessage)
		if !requestHTTPMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(requestHTTPMessage, &responseHTTPMessage, protoMinor)
	}

	matcher.openMessagesMap.Store(ident, &responseHTTPMessage)
	return nil
}

func (matcher *requestResponseMatcher) preparePair(requestHTTPMessage *api.GenericMessage, responseHTTPMessage *api.GenericMessage, protoMinor int) *api.OutputChannelItem {
	protocol := http11protocol
	if protoMinor == 0 {
		protocol = http10protocol
	}
	return &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      requestHTTPMessage.CaptureTime.UnixNano() / int64(time.Millisecond),
		ConnectionInfo: nil,
		Pair: &api.RequestResponsePair{
			Request:  *requestHTTPMessage,
			Response: *responseHTTPMessage,
		},
	}
}
