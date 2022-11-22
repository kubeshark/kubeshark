package amqp

import (
	"sync"
	"time"

	"github.com/kubeshark/kubeshark/tap/api"
)

// Key is {client_addr}_{client_port}_{dest_addr}_{dest_port}_{channel_id}_{class_id}_{method_id}
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

func (matcher *requestResponseMatcher) emitEvent(isRequest bool, ident string, method string, event interface{}, reader api.TcpReader) {
	reader.GetParent().SetProtocol(&protocol)

	var item *api.OutputChannelItem
	if isRequest {
		item = matcher.registerRequest(ident, method, event, reader.GetCaptureTime(), reader.GetReadProgress().Current())
	} else {
		item = matcher.registerResponse(ident, method, event, reader.GetCaptureTime(), reader.GetReadProgress().Current())
	}

	if item != nil {
		item.ConnectionInfo = &api.ConnectionInfo{
			ClientIP:   reader.GetTcpID().SrcIP,
			ClientPort: reader.GetTcpID().SrcPort,
			ServerIP:   reader.GetTcpID().DstIP,
			ServerPort: reader.GetTcpID().DstPort,
			IsOutgoing: true,
		}
		item.Capture = reader.GetParent().GetOrigin()
		reader.GetEmitter().Emit(item)
	}
}

func (matcher *requestResponseMatcher) registerRequest(ident string, method string, request interface{}, captureTime time.Time, captureSize int) *api.OutputChannelItem {
	requestAMQPMessage := api.GenericMessage{
		IsRequest:   true,
		CaptureTime: captureTime,
		CaptureSize: captureSize,
		Payload: AMQPPayload{
			Data: &AMQPWrapper{
				Method:  method,
				Url:     "",
				Details: request,
			},
		},
	}

	if response, found := matcher.openMessagesMap.LoadAndDelete(ident); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		responseAMQPMessage := response.(*api.GenericMessage)
		if responseAMQPMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(&requestAMQPMessage, responseAMQPMessage)
	}

	matcher.openMessagesMap.Store(ident, &requestAMQPMessage)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(ident string, method string, response interface{}, captureTime time.Time, captureSize int) *api.OutputChannelItem {
	responseAMQPMessage := api.GenericMessage{
		IsRequest:   false,
		CaptureTime: captureTime,
		CaptureSize: captureSize,
		Payload: AMQPPayload{
			Data: &AMQPWrapper{
				Method:  method,
				Url:     "",
				Details: response,
			},
		},
	}

	if request, found := matcher.openMessagesMap.LoadAndDelete(ident); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		requestAMQPMessage := request.(*api.GenericMessage)
		if !requestAMQPMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(requestAMQPMessage, &responseAMQPMessage)
	}

	matcher.openMessagesMap.Store(ident, &responseAMQPMessage)
	return nil
}

func (matcher *requestResponseMatcher) preparePair(requestAMQPMessage *api.GenericMessage, responseAMQPMessage *api.GenericMessage) *api.OutputChannelItem {
	return &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      requestAMQPMessage.CaptureTime.UnixNano() / int64(time.Millisecond),
		ConnectionInfo: nil,
		Pair: &api.RequestResponsePair{
			Request:  *requestAMQPMessage,
			Response: *responseAMQPMessage,
		},
	}
}
