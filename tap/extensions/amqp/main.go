package amqp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

var protocol api.Protocol = api.Protocol{
	Name:            "amqp",
	LongName:        "Advanced Message Queuing Protocol 0-9-1",
	Abbreviation:    "AMQP",
	Macro:           "amqp",
	Version:         "0-9-1",
	BackgroundColor: "#ff6600",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://www.rabbitmq.com/amqp-0-9-1-reference.html",
	Ports:           []string{"5671", "5672"},
	Priority:        1,
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s", protocol.Name)
}

const amqpRequest string = "amqp_request"

func (d dissecting) Dissect(b *bufio.Reader, reader api.TcpReader, options *api.TrafficFilteringOptions) error {
	r := AmqpReader{b}

	var remaining int
	var header *HeaderFrame

	connectionInfo := &api.ConnectionInfo{
		ClientIP:   reader.GetTcpID().SrcIP,
		ClientPort: reader.GetTcpID().SrcPort,
		ServerIP:   reader.GetTcpID().DstIP,
		ServerPort: reader.GetTcpID().DstPort,
		IsOutgoing: true,
	}

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

	for {
		frame, err := r.ReadFrame()
		if err == io.EOF {
			// We must read until we see an EOF... very important!
			return err
		}

		switch f := frame.(type) {
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
				emitAMQP(*eventBasicPublish, amqpRequest, basicMethodMap[40], connectionInfo, reader.GetCaptureTime(), reader.GetReadProgress().Current(), reader.GetEmitter(), reader.GetParent().GetOrigin())
			case *BasicDeliver:
				eventBasicDeliver.Body = f.Body
				emitAMQP(*eventBasicDeliver, amqpRequest, basicMethodMap[60], connectionInfo, reader.GetCaptureTime(), reader.GetReadProgress().Current(), reader.GetEmitter(), reader.GetParent().GetOrigin())
			}

		case *MethodFrame:
			reader.GetParent().SetProtocol(&protocol)

			lastMethodFrameMessage = f.Method
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
				emitAMQP(*eventQueueBind, amqpRequest, queueMethodMap[20], connectionInfo, reader.GetCaptureTime(), reader.GetReadProgress().Current(), reader.GetEmitter(), reader.GetParent().GetOrigin())

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
				emitAMQP(*eventBasicConsume, amqpRequest, basicMethodMap[20], connectionInfo, reader.GetCaptureTime(), reader.GetReadProgress().Current(), reader.GetEmitter(), reader.GetParent().GetOrigin())

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
				emitAMQP(*eventQueueDeclare, amqpRequest, queueMethodMap[10], connectionInfo, reader.GetCaptureTime(), reader.GetReadProgress().Current(), reader.GetEmitter(), reader.GetParent().GetOrigin())

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
				emitAMQP(*eventExchangeDeclare, amqpRequest, exchangeMethodMap[10], connectionInfo, reader.GetCaptureTime(), reader.GetReadProgress().Current(), reader.GetEmitter(), reader.GetParent().GetOrigin())

			case *ConnectionStart:
				eventConnectionStart := &ConnectionStart{
					VersionMajor:     m.VersionMajor,
					VersionMinor:     m.VersionMinor,
					ServerProperties: m.ServerProperties,
					Mechanisms:       m.Mechanisms,
					Locales:          m.Locales,
				}
				emitAMQP(*eventConnectionStart, amqpRequest, connectionMethodMap[10], connectionInfo, reader.GetCaptureTime(), reader.GetReadProgress().Current(), reader.GetEmitter(), reader.GetParent().GetOrigin())

			case *ConnectionClose:
				eventConnectionClose := &ConnectionClose{
					ReplyCode: m.ReplyCode,
					ReplyText: m.ReplyText,
					ClassId:   m.ClassId,
					MethodId:  m.MethodId,
				}
				emitAMQP(*eventConnectionClose, amqpRequest, connectionMethodMap[50], connectionInfo, reader.GetCaptureTime(), reader.GetReadProgress().Current(), reader.GetEmitter(), reader.GetParent().GetOrigin())
			}

		default:
			// log.Printf("unexpected frame: %+v", f)
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, resolvedSource string, resolvedDestination string, namespace string) *api.Entry {
	request := item.Pair.Request.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})

	reqDetails["method"] = request["method"]
	return &api.Entry{
		Protocol: protocol,
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
		Namespace:   namespace,
		Outgoing:    item.ConnectionInfo.IsOutgoing,
		Request:     reqDetails,
		RequestSize: item.Pair.Request.CaptureSize,
		Timestamp:   item.Timestamp,
		StartTime:   item.Pair.Request.CaptureTime,
		ElapsedTime: 0,
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
	}

	return &api.BaseEntry{
		Id:             entry.Id,
		Protocol:       entry.Protocol,
		Capture:        entry.Capture,
		Summary:        summary,
		SummaryQuery:   summaryQuery,
		Status:         0,
		StatusQuery:    "",
		Method:         method,
		MethodQuery:    methodQuery,
		Timestamp:      entry.Timestamp,
		Source:         entry.Source,
		Destination:    entry.Destination,
		IsOutgoing:     entry.Outgoing,
		Latency:        entry.ElapsedTime,
		Rules:          entry.Rules,
		ContractStatus: entry.ContractStatus,
	}
}

func (d dissecting) Represent(request map[string]interface{}, response map[string]interface{}) (object []byte, err error) {
	representation := make(map[string]interface{})
	var repRequest []interface{}
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
	}
	representation["request"] = repRequest
	object, err = json.Marshal(representation)
	return
}

func (d dissecting) Macros() map[string]string {
	return map[string]string{
		`amqp`: fmt.Sprintf(`proto.name == "%s"`, protocol.Name),
	}
}

func (d dissecting) NewResponseRequestMatcher() api.RequestResponseMatcher {
	return nil
}

var Dissector dissecting

func NewDissector() api.Dissector {
	return Dissector
}
