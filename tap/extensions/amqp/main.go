package amqp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/kubeshark/kubeshark/tap/api"
)

var protocol = api.Protocol{
	ProtocolSummary: api.ProtocolSummary{
		Name:         "amqp",
		Version:      "0-9-1",
		Abbreviation: "AMQP",
	},
	LongName:        "Advanced Message Queuing Protocol 0-9-1",
	Macro:           "amqp",
	BackgroundColor: "#ff6600",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://www.rabbitmq.com/amqp-0-9-1-reference.html",
	Ports:           []string{"5671", "5672"},
	Priority:        1,
}

var protocolsMap = map[string]*api.Protocol{
	protocol.ToString(): &protocol,
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &protocol
}

func (d dissecting) GetProtocols() map[string]*api.Protocol {
	return protocolsMap
}

func (d dissecting) Ping() {
	log.Printf("pong %s", protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, reader api.TcpReader, options *api.TrafficFilteringOptions) error {
	r := AmqpReader{b}

	var remaining int
	var header *HeaderFrame

	eventBasicPublish := &BasicPublish{
		Exchange:   "",
		RoutingKey: "",
		Mandatory:  false,
		Immediate:  false,
		Body:       nil,
		Properties: Properties{},
	}

	eventBasicDeliver := &BasicDeliver{
		ConsumerTag: "",
		DeliveryTag: 0,
		Redelivered: false,
		Exchange:    "",
		RoutingKey:  "",
		Properties:  Properties{},
		Body:        nil,
	}

	var lastMethodFrameMessage Message

	var ident string
	isClient := reader.GetIsClient()
	reqResMatcher := reader.GetReqResMatcher().(*requestResponseMatcher)

	for {
		frameVal, err := r.readFrame()
		if err == io.EOF {
			// We must read until we see an EOF... very important!
			return err
		}

		switch f := frameVal.(type) {
		case *HeartbeatFrame:
			// drop

		case *HeaderFrame:
			reader.GetParent().SetProtocol(&protocol)

			// start content state
			header = f
			remaining = int(header.Size)

			// Workaround for `Time.MarshalJSON: year outside of range [0,9999]` error
			if header.Properties.Timestamp.Year() > 9999 {
				header.Properties.Timestamp = time.Time{}.UTC()
			}

			switch lastMethodFrameMessage.(type) {
			case *BasicPublish:
				eventBasicPublish.Properties = header.Properties
			case *BasicDeliver:
				eventBasicDeliver.Properties = header.Properties
			}

		case *BodyFrame:
			reader.GetParent().SetProtocol(&protocol)

			// continue until terminated
			remaining -= len(f.Body)
			switch lastMethodFrameMessage.(type) {
			case *BasicPublish:
				eventBasicPublish.Body = f.Body
				reqResMatcher.emitEvent(isClient, ident, basicMethodMap[40], *eventBasicPublish, reader)
				reqResMatcher.emitEvent(!isClient, ident, emptyMethod, &emptyResponse{}, reader)

			case *BasicDeliver:
				eventBasicDeliver.Body = f.Body
				reqResMatcher.emitEvent(!isClient, ident, basicMethodMap[60], *eventBasicDeliver, reader)
				reqResMatcher.emitEvent(isClient, ident, emptyMethod, &emptyResponse{}, reader)
			}

		case *MethodFrame:
			reader.GetParent().SetProtocol(&protocol)

			lastMethodFrameMessage = f.Method

			ident = getIdent(reader, f)

			switch m := f.Method.(type) {
			case *BasicPublish:
				eventBasicPublish.Exchange = m.Exchange
				eventBasicPublish.RoutingKey = m.RoutingKey
				eventBasicPublish.Mandatory = m.Mandatory
				eventBasicPublish.Immediate = m.Immediate

			case *QueueBind:
				eventQueueBind := &QueueBind{
					Queue:      m.Queue,
					Exchange:   m.Exchange,
					RoutingKey: m.RoutingKey,
					NoWait:     m.NoWait,
					Arguments:  m.Arguments,
				}
				reqResMatcher.emitEvent(isClient, ident, queueMethodMap[20], *eventQueueBind, reader)

			case *QueueBindOk:
				reqResMatcher.emitEvent(isClient, ident, queueMethodMap[21], m, reader)

			case *BasicConsume:
				eventBasicConsume := &BasicConsume{
					Queue:       m.Queue,
					ConsumerTag: m.ConsumerTag,
					NoLocal:     m.NoLocal,
					NoAck:       m.NoAck,
					Exclusive:   m.Exclusive,
					NoWait:      m.NoWait,
					Arguments:   m.Arguments,
				}
				reqResMatcher.emitEvent(isClient, ident, basicMethodMap[20], *eventBasicConsume, reader)

			case *BasicConsumeOk:
				reqResMatcher.emitEvent(isClient, ident, basicMethodMap[21], m, reader)

			case *BasicDeliver:
				eventBasicDeliver.ConsumerTag = m.ConsumerTag
				eventBasicDeliver.DeliveryTag = m.DeliveryTag
				eventBasicDeliver.Redelivered = m.Redelivered
				eventBasicDeliver.Exchange = m.Exchange
				eventBasicDeliver.RoutingKey = m.RoutingKey

			case *QueueDeclare:
				eventQueueDeclare := &QueueDeclare{
					Queue:      m.Queue,
					Passive:    m.Passive,
					Durable:    m.Durable,
					AutoDelete: m.AutoDelete,
					Exclusive:  m.Exclusive,
					NoWait:     m.NoWait,
					Arguments:  m.Arguments,
				}
				reqResMatcher.emitEvent(isClient, ident, queueMethodMap[10], *eventQueueDeclare, reader)

			case *QueueDeclareOk:
				reqResMatcher.emitEvent(isClient, ident, queueMethodMap[11], m, reader)

			case *ExchangeDeclare:
				eventExchangeDeclare := &ExchangeDeclare{
					Exchange:   m.Exchange,
					Type:       m.Type,
					Passive:    m.Passive,
					Durable:    m.Durable,
					AutoDelete: m.AutoDelete,
					Internal:   m.Internal,
					NoWait:     m.NoWait,
					Arguments:  m.Arguments,
				}
				reqResMatcher.emitEvent(isClient, ident, exchangeMethodMap[10], *eventExchangeDeclare, reader)

			case *ExchangeDeclareOk:
				reqResMatcher.emitEvent(isClient, ident, exchangeMethodMap[11], m, reader)

			case *ConnectionStart:
				// In our tests, *ConnectionStart does not result in *ConnectionStartOk
				reqResMatcher.emitEvent(!isClient, ident, connectionMethodMap[10], m, reader)
				reqResMatcher.emitEvent(isClient, ident, emptyMethod, &emptyResponse{}, reader)

			case *ConnectionStartOk:
				// In our tests, *ConnectionStart does not result in *ConnectionStartOk
				reqResMatcher.emitEvent(isClient, ident, connectionMethodMap[11], m, reader)

			case *ConnectionClose:
				eventConnectionClose := &ConnectionClose{
					ReplyCode: m.ReplyCode,
					ReplyText: m.ReplyText,
					ClassId:   m.ClassId,
					MethodId:  m.MethodId,
				}
				reqResMatcher.emitEvent(isClient, ident, connectionMethodMap[50], *eventConnectionClose, reader)

			case *ConnectionCloseOk:
				reqResMatcher.emitEvent(isClient, ident, connectionMethodMap[51], m, reader)

			case *connectionOpen:
				eventConnectionOpen := &connectionOpen{
					VirtualHost: m.VirtualHost,
				}
				reqResMatcher.emitEvent(isClient, ident, connectionMethodMap[40], *eventConnectionOpen, reader)

			case *connectionOpenOk:
				reqResMatcher.emitEvent(isClient, ident, connectionMethodMap[41], m, reader)

			case *channelOpen:
				reqResMatcher.emitEvent(isClient, ident, channelMethodMap[10], m, reader)

			case *channelOpenOk:
				reqResMatcher.emitEvent(isClient, ident, channelMethodMap[11], m, reader)

			case *connectionTune:
				// In our tests, *connectionTune does not result in *connectionTuneOk
				reqResMatcher.emitEvent(!isClient, ident, connectionMethodMap[30], m, reader)
				reqResMatcher.emitEvent(isClient, ident, emptyMethod, &emptyResponse{}, reader)

			case *connectionTuneOk:
				// In our tests, *connectionTune does not result in *connectionTuneOk
				reqResMatcher.emitEvent(isClient, ident, connectionMethodMap[31], m, reader)

			case *basicCancel:
				reqResMatcher.emitEvent(isClient, ident, basicMethodMap[30], m, reader)

			case *basicCancelOk:
				reqResMatcher.emitEvent(isClient, ident, basicMethodMap[31], m, reader)
			}

		default:
			// log.Printf("unexpected frameVal: %+v", f)
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, resolvedSource string, resolvedDestination string, namespace string) *api.Entry {
	request := item.Pair.Request.Payload.(map[string]interface{})
	response := item.Pair.Response.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	resDetails := response["details"].(map[string]interface{})

	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	if elapsedTime < 0 {
		elapsedTime = 0
	}

	reqDetails["method"] = request["method"]
	resDetails["method"] = response["method"]
	return &api.Entry{
		Protocol: protocol.ProtocolSummary,
		Capture:  item.Capture,
		Source: &api.TCP{
			Name: resolvedSource,
			IP:   item.ConnectionInfo.ClientIP,
			Port: item.ConnectionInfo.ClientPort,
		},
		Destination: &api.TCP{
			Name: resolvedDestination,
			IP:   item.ConnectionInfo.ServerIP,
			Port: item.ConnectionInfo.ServerPort,
		},
		Namespace:    namespace,
		Outgoing:     item.ConnectionInfo.IsOutgoing,
		Request:      reqDetails,
		Response:     resDetails,
		RequestSize:  item.Pair.Request.CaptureSize,
		ResponseSize: item.Pair.Response.CaptureSize,
		Timestamp:    item.Timestamp,
		StartTime:    item.Pair.Request.CaptureTime,
		ElapsedTime:  elapsedTime,
	}

}

func (d dissecting) Summarize(entry *api.Entry) *api.BaseEntry {
	summary := ""
	summaryQuery := ""
	method := entry.Request["method"].(string)
	methodQuery := fmt.Sprintf(`request.method == "%s"`, method)
	switch method {
	case basicMethodMap[40]:
		summary = entry.Request["exchange"].(string)
		summaryQuery = fmt.Sprintf(`request.exchange == "%s"`, summary)
	case basicMethodMap[60]:
		summary = entry.Request["exchange"].(string)
		summaryQuery = fmt.Sprintf(`request.exchange == "%s"`, summary)
	case exchangeMethodMap[10]:
		summary = entry.Request["exchange"].(string)
		summaryQuery = fmt.Sprintf(`request.exchange == "%s"`, summary)
	case queueMethodMap[10]:
		summary = entry.Request["queue"].(string)
		summaryQuery = fmt.Sprintf(`request.queue == "%s"`, summary)
	case connectionMethodMap[10]:
		versionMajor := int(entry.Request["versionMajor"].(float64))
		versionMinor := int(entry.Request["versionMinor"].(float64))
		summary = fmt.Sprintf(
			"%s.%s",
			strconv.Itoa(versionMajor),
			strconv.Itoa(versionMinor),
		)
		summaryQuery = fmt.Sprintf(`request.versionMajor == %d and request.versionMinor == %d`, versionMajor, versionMinor)
	case connectionMethodMap[50]:
		summary = entry.Request["replyText"].(string)
		summaryQuery = fmt.Sprintf(`request.replyText == "%s"`, summary)
	case queueMethodMap[20]:
		summary = entry.Request["queue"].(string)
		summaryQuery = fmt.Sprintf(`request.queue == "%s"`, summary)
	case basicMethodMap[20]:
		summary = entry.Request["queue"].(string)
		summaryQuery = fmt.Sprintf(`request.queue == "%s"`, summary)
	case connectionMethodMap[40]:
		summary = entry.Request["virtualHost"].(string)
		summaryQuery = fmt.Sprintf(`request.virtualHost == "%s"`, summary)
	case connectionMethodMap[30]:
		summary = fmt.Sprintf("%g", entry.Request["channelMax"].(float64))
		summaryQuery = fmt.Sprintf(`request.channelMax == "%s"`, summary)
	case connectionMethodMap[31]:
		summary = fmt.Sprintf("%g", entry.Request["channelMax"].(float64))
		summaryQuery = fmt.Sprintf(`request.channelMax == "%s"`, summary)
	case basicMethodMap[30]:
		summary = entry.Request["consumerTag"].(string)
		summaryQuery = fmt.Sprintf(`request.consumerTag == "%s"`, summary)
	case basicMethodMap[31]:
		summary = entry.Request["consumerTag"].(string)
		summaryQuery = fmt.Sprintf(`request.consumerTag == "%s"`, summary)
	}

	return &api.BaseEntry{
		Id:           entry.Id,
		Protocol:     *protocolsMap[entry.Protocol.ToString()],
		Capture:      entry.Capture,
		Summary:      summary,
		SummaryQuery: summaryQuery,
		Status:       0,
		StatusQuery:  "",
		Method:       method,
		MethodQuery:  methodQuery,
		Timestamp:    entry.Timestamp,
		Source:       entry.Source,
		Destination:  entry.Destination,
		IsOutgoing:   entry.Outgoing,
		Latency:      entry.ElapsedTime,
	}
}

func (d dissecting) Represent(request map[string]interface{}, response map[string]interface{}) (object []byte, err error) {
	representation := make(map[string]interface{})
	var repRequest []interface{}
	var repResponse []interface{}

	switch request["method"].(string) {
	case basicMethodMap[40]:
		repRequest = representBasicPublish(request)
	case basicMethodMap[60]:
		repRequest = representBasicDeliver(request)
	case queueMethodMap[10]:
		repRequest = representQueueDeclare(request)
	case exchangeMethodMap[10]:
		repRequest = representExchangeDeclare(request)
	case connectionMethodMap[10]:
		repRequest = representConnectionStart(request)
	case connectionMethodMap[50]:
		repRequest = representConnectionClose(request)
	case queueMethodMap[20]:
		repRequest = representQueueBind(request)
	case basicMethodMap[20]:
		repRequest = representBasicConsume(request)
	case connectionMethodMap[40]:
		repRequest = representConnectionOpen(request)
	case channelMethodMap[10]:
		repRequest = representEmpty(request)
	case connectionMethodMap[30]:
		repRequest = representConnectionTune(request)
	case basicMethodMap[30]:
		repRequest = representBasicCancel(request)
	}

	switch response["method"].(string) {
	case queueMethodMap[11]:
		repResponse = representQueueDeclareOk(response)
	case exchangeMethodMap[11]:
		repResponse = representEmpty(response)
	case connectionMethodMap[11]:
		repResponse = representConnectionStartOk(response)
	case connectionMethodMap[51]:
		repResponse = representEmpty(response)
	case basicMethodMap[21]:
		repResponse = representBasicConsumeOk(response)
	case queueMethodMap[21]:
		repResponse = representEmpty(response)
	case connectionMethodMap[41]:
		repResponse = representEmpty(response)
	case channelMethodMap[11]:
		repResponse = representEmpty(request)
	case connectionMethodMap[31]:
		repResponse = representConnectionTune(request)
	case basicMethodMap[31]:
		repResponse = representBasicCancelOk(request)
	case emptyMethod:
		repResponse = representEmpty(response)
	}

	representation["request"] = repRequest
	representation["response"] = repResponse
	object, err = json.Marshal(representation)

	return
}

func (d dissecting) Macros() map[string]string {
	return map[string]string{
		`amqp`: fmt.Sprintf(`protocol.name == "%s"`, protocol.Name),
	}
}

func (d dissecting) NewResponseRequestMatcher() api.RequestResponseMatcher {
	return createResponseRequestMatcher()
}

var Dissector dissecting

func NewDissector() api.Dissector {
	return Dissector
}
