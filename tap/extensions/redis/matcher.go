package redis

import (
	"sync"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

// Key is `{src_ip}_{dst_ip}_{src_ip}_{src_port}_{incremental_counter}`
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

func (matcher *requestResponseMatcher) registerRequest(ident string, request *RedisPacket, captureTime time.Time, captureSize int) *api.OutputChannelItem {
	requestRedisMessage := api.GenericMessage{
		IsRequest:   true,
		CaptureTime: captureTime,
		CaptureSize: captureSize,
		Payload: RedisPayload{
			Data: &RedisWrapper{
				Method:  string(request.Command),
				Url:     "",
				Details: request,
			},
		},
	}

	if response, found := matcher.openMessagesMap.LoadAndDelete(ident); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		responseRedisMessage := response.(*api.GenericMessage)
		if responseRedisMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(&requestRedisMessage, responseRedisMessage)
	}

	matcher.openMessagesMap.Store(ident, &requestRedisMessage)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(ident string, response *RedisPacket, captureTime time.Time, captureSize int) *api.OutputChannelItem {
	responseRedisMessage := api.GenericMessage{
		IsRequest:   false,
		CaptureTime: captureTime,
		CaptureSize: captureSize,
		Payload: RedisPayload{
			Data: &RedisWrapper{
				Method:  string(response.Command),
				Url:     "",
				Details: response,
			},
		},
	}

	if request, found := matcher.openMessagesMap.LoadAndDelete(ident); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		requestRedisMessage := request.(*api.GenericMessage)
		if !requestRedisMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(requestRedisMessage, &responseRedisMessage)
	}

	matcher.openMessagesMap.Store(ident, &responseRedisMessage)
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
