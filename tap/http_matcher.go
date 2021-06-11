package tap

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/orcaman/concurrent-map"
)

type requestResponsePair struct {
	Request  httpMessage `json:"request"`
	Response httpMessage `json:"response"`
}

type Connection struct {
	ClientIP   string
	ClientPort string
	ServerIP   string
	ServerPort string
}

type httpMessage struct {
	isRequest       bool
	captureTime     time.Time
	orig            interface {}
	connection       Connection
}


// Key is {client_addr}:{client_port}->{dest_addr}:{dest_port}
type requestResponseMatcher struct {
	openMessagesMap cmap.ConcurrentMap

}

func createResponseRequestMatcher() requestResponseMatcher {
	newMatcher := &requestResponseMatcher{openMessagesMap: cmap.New()}
	return *newMatcher
}

func (matcher *requestResponseMatcher) registerRequest(ident string, request *http.Request, captureTime time.Time) *requestResponsePair {
	split := splitIdent(ident)
	key := genKey(split)

	connection := &Connection{
		ClientIP:   split[0],
		ClientPort: split[2],
		ServerIP:   split[1],
		ServerPort: split[3],
	}

	requestHTTPMessage := httpMessage{
		isRequest:       true,
		captureTime:     captureTime,
		orig:            request,
		connection:      *connection,
	}

	if response, found := matcher.openMessagesMap.Pop(key); found {
		// Type assertion always succeeds because all of the map's values are of httpMessage type
		responseHTTPMessage := response.(*httpMessage)
		if responseHTTPMessage.isRequest {
			SilentError("Request-Duplicate", "Got duplicate request with same identifier")
			return nil
		}
		Debug("Matched open Response for %s", key)
		return matcher.preparePair(&requestHTTPMessage, responseHTTPMessage)
	}

	matcher.openMessagesMap.Set(key, &requestHTTPMessage)
	Debug("Registered open Request for %s", key)
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

	if request, found := matcher.openMessagesMap.Pop(key); found {
		// Type assertion always succeeds because all of the map's values are of httpMessage type
		requestHTTPMessage := request.(*httpMessage)
		if !requestHTTPMessage.isRequest {
			SilentError("Response-Duplicate", "Got duplicate response with same identifier")
			return nil
		}
		Debug("Matched open Request for %s", key)
		return matcher.preparePair(requestHTTPMessage, &responseHTTPMessage)
	}

	matcher.openMessagesMap.Set(key, &responseHTTPMessage)
	Debug("Registered open Response for %s", key)
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
