package redis

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

var reqResMatcher = createResponseRequestMatcher() // global

// Key is {client_addr}:{client_port}->{dest_addr}:{dest_port}_{incremental_counter}
type requestResponseMatcher struct {
	openMessagesMap *sync.Map
}

func createResponseRequestMatcher() requestResponseMatcher {
	newMatcher := &requestResponseMatcher{openMessagesMap: &sync.Map{}}
	return *newMatcher
}

func (matcher *requestResponseMatcher) registerRequest(ident string, request *RedisPacket, captureTime time.Time) *api.OutputChannelItem {
	split := splitIdent(ident)
	key := genKey(split)

	requestRedisMessage := api.GenericMessage{
		IsRequest:   true,
		CaptureTime: captureTime,
		Payload: RedisPayload{
			Data: &RedisWrapper{
				Method:  string(request.Command),
				Url:     "",
				Details: request,
			},
		},
	}

	if response, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		responseRedisMessage := response.(*api.GenericMessage)
		if responseRedisMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(&requestRedisMessage, responseRedisMessage)
	}

	matcher.openMessagesMap.Store(key, &requestRedisMessage)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(ident string, response *RedisPacket, captureTime time.Time) *api.OutputChannelItem {
	split := splitIdent(ident)
	key := genKey(split)

	responseRedisMessage := api.GenericMessage{
		IsRequest:   false,
		CaptureTime: captureTime,
		Payload: RedisPayload{
			Data: &RedisWrapper{
				Method:  string(response.Command),
				Url:     "",
				Details: response,
			},
		},
	}

	if request, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		requestRedisMessage := request.(*api.GenericMessage)
		if !requestRedisMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(requestRedisMessage, &responseRedisMessage)
	}

	matcher.openMessagesMap.Store(key, &responseRedisMessage)
	return nil
}

func (matcher *requestResponseMatcher) preparePair(requestRedisMessage *api.GenericMessage, responseRedisMessage *api.GenericMessage) *api.OutputChannelItem {
	return &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      requestRedisMessage.CaptureTime.UnixNano() / int64(time.Millisecond),
		ConnectionInfo: nil,
		Pair: &api.RequestResponsePair{
			Request:  *requestRedisMessage,
			Response: *responseRedisMessage,
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
