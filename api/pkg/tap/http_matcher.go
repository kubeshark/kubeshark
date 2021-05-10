package tap

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/orcaman/concurrent-map"
)

type requestResponsePair struct {
	Request  httpMessage `json:"request"`
	Response httpMessage `json:"response"`
}

type envoyMessageWrapper struct {
	HttpBufferedTrace requestResponsePair `json:"http_buffered_trace"`
}

type headerKeyVal struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type messageBody struct {
	Truncated bool   `json:"truncated"`
	AsBytes   string `json:"as_bytes"`
}

type httpMessage struct {
	IsRequest       bool
	Headers         []headerKeyVal `json:"headers"`
	HTTPVersion     string         `json:"httpVersion"`
	Body            messageBody    `json:"body"`
	captureTime     time.Time
	orig            interface {}
	requestSenderIp string
}


// Key is {client_addr}:{client_port}->{dest_addr}:{dest_port}
type requestResponseMatcher struct {
	openMessagesMap cmap.ConcurrentMap

}

func createResponseRequestMatcher() requestResponseMatcher {
	newMatcher := &requestResponseMatcher{openMessagesMap: cmap.New()}
	return *newMatcher
}

func (matcher *requestResponseMatcher) registerRequest(ident string, request *http.Request, captureTime time.Time, body string, isHTTP2 bool) *envoyMessageWrapper {
	split := splitIdent(ident)
	key := genKey(split)

	messageExtraHeaders := []headerKeyVal{
		{Key: "x-up9-source", Value: split[0]},
		{Key: "x-up9-destination", Value: split[1] + ":" + split[3]},
	}

	requestHTTPMessage := requestToMessage(request, captureTime, body, &messageExtraHeaders, isHTTP2, split[0])

	if response, found := matcher.openMessagesMap.Pop(key); found {
		// Type assertion always succeeds because all of the map's values are of httpMessage type
		responseHTTPMessage := response.(*httpMessage)
		if responseHTTPMessage.IsRequest {
			SilentError("Request-Duplicate", "Got duplicate request with same identifier\n")
			return nil
		}
		Debug("Matched open Response for %s\n", key)
		return matcher.preparePair(&requestHTTPMessage, responseHTTPMessage)
	}

	matcher.openMessagesMap.Set(key, &requestHTTPMessage)
	Debug("Registered open Request for %s\n", key)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(ident string, response *http.Response, captureTime time.Time, body string, isHTTP2 bool) *envoyMessageWrapper {
	split := splitIdent(ident)
	key := genKey(split)

	responseHTTPMessage := responseToMessage(response, captureTime, body, isHTTP2)

	if request, found := matcher.openMessagesMap.Pop(key); found {
		// Type assertion always succeeds because all of the map's values are of httpMessage type
		requestHTTPMessage := request.(*httpMessage)
		if !requestHTTPMessage.IsRequest {
			SilentError("Response-Duplicate", "Got duplicate response with same identifier\n")
			return nil
		}
		Debug("Matched open Request for %s\n", key)
		return matcher.preparePair(requestHTTPMessage, &responseHTTPMessage)
	}

	matcher.openMessagesMap.Set(key, &responseHTTPMessage)
	Debug("Registered open Response for %s\n", key)
	return nil
}

func (matcher *requestResponseMatcher) preparePair(requestHTTPMessage *httpMessage, responseHTTPMessage *httpMessage) *envoyMessageWrapper {
	matcher.addDuration(requestHTTPMessage, responseHTTPMessage)

	return &envoyMessageWrapper{
		HttpBufferedTrace: requestResponsePair{
			Request:  *requestHTTPMessage,
			Response: *responseHTTPMessage,
		},
	}
}

func requestToMessage(request *http.Request, captureTime time.Time, body string, messageExtraHeaders *[]headerKeyVal, isHTTP2 bool, requestSenderIp string) httpMessage {
	messageHeaders := make([]headerKeyVal, 0)

	for key, value := range request.Header {
		messageHeaders = append(messageHeaders, headerKeyVal{Key: key, Value: value[0]})
	}

	if !isHTTP2 {
		messageHeaders = append(messageHeaders, headerKeyVal{Key: ":method", Value: request.Method})
		messageHeaders = append(messageHeaders, headerKeyVal{Key: ":path", Value: request.RequestURI})
		messageHeaders = append(messageHeaders, headerKeyVal{Key: ":authority", Value: request.Host})
		messageHeaders = append(messageHeaders, headerKeyVal{Key: ":scheme", Value: "http"})
	}

	messageHeaders = append(messageHeaders, headerKeyVal{Key: "x-request-start", Value: fmt.Sprintf("%.3f", float64(captureTime.UnixNano()) / float64(1000000000))})

	messageHeaders = append(messageHeaders, *messageExtraHeaders...)

	httpVersion := request.Proto

	requestBody := messageBody{Truncated: false, AsBytes: body}

	return httpMessage{
		IsRequest:       true,
		Headers:         messageHeaders,
		HTTPVersion:     httpVersion,
		Body:            requestBody,
		captureTime:     captureTime,
		orig:            request,
		requestSenderIp: requestSenderIp,
	}
}

func responseToMessage(response *http.Response, captureTime time.Time, body string, isHTTP2 bool) httpMessage {
	messageHeaders := make([]headerKeyVal, 0)

	for key, value := range response.Header {
		messageHeaders = append(messageHeaders, headerKeyVal{Key: key, Value: value[0]})
	}

	if !isHTTP2 {
		messageHeaders = append(messageHeaders, headerKeyVal{Key: ":status", Value: strconv.Itoa(response.StatusCode)})
	}

	httpVersion := response.Proto

	requestBody := messageBody{Truncated: false, AsBytes: body}

	return httpMessage{
		IsRequest:   false,
		Headers:     messageHeaders,
		HTTPVersion: httpVersion,
		Body:        requestBody,
		captureTime: captureTime,
		orig:        response,
	}
}

func (matcher *requestResponseMatcher) addDuration(requestHTTPMessage *httpMessage, responseHTTPMessage *httpMessage) {
	durationMs := float64(responseHTTPMessage.captureTime.UnixNano() / 1000000) - float64(requestHTTPMessage.captureTime.UnixNano() / 1000000)
	if durationMs < 1 {
		durationMs = 1
	}

	responseHTTPMessage.Headers  = append(responseHTTPMessage.Headers, headerKeyVal{Key: "x-up9-duration-ms", Value: fmt.Sprintf("%.0f", durationMs)})
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
	keysToPop := make([]string, 0)
	for item := range matcher.openMessagesMap.IterBuffered() {
		// Map only contains values of type httpMessage
		message, _ := item.Val.(*httpMessage)

		if message.captureTime.Before(t) {
			keysToPop = append(keysToPop, item.Key)
		}
	}

	numDeleted := len(keysToPop)
	
	for _, key := range keysToPop {
		_, _ = matcher.openMessagesMap.Pop(key)
	}

	return numDeleted
}
