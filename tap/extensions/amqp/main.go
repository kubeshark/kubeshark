package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"

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

func init() {
	log.Println("Initializing AMQP extension...")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s\n", protocol.Name)
}

const amqpRequest string = "amqp_request"

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, superIdentifier *api.SuperIdentifier, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	r := AmqpReader{b}

	var remaining int
	var header *HeaderFrame
	var body []byte

	connectionInfo := &api.ConnectionInfo{
		ClientIP:   tcpID.SrcIP,
		ClientPort: tcpID.SrcPort,
		ServerIP:   tcpID.DstIP,
		ServerPort: tcpID.DstPort,
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
		if superIdentifier.Protocol != nil && superIdentifier.Protocol != &protocol {
			return errors.New("Identified by another protocol")
		}

		frame, err := r.ReadFrame()
		if err == io.EOF {
			// We must read until we see an EOF... very important!
			return errors.New("AMQP EOF")
		}

		switch f := frame.(type) {
		case *HeartbeatFrame:
			// drop

		case *HeaderFrame:
			// start content state
			header = f
			remaining = int(header.Size)
			switch lastMethodFrameMessage.(type) {
			case *BasicPublish:
				eventBasicPublish.Properties = header.Properties
			case *BasicDeliver:
				eventBasicDeliver.Properties = header.Properties
			default:
				frame = nil
			}

		case *BodyFrame:
			// continue until terminated
			body = append(body, f.Body...)
			remaining -= len(f.Body)
			switch lastMethodFrameMessage.(type) {
			case *BasicPublish:
				eventBasicPublish.Body = f.Body
				superIdentifier.Protocol = &protocol
				emitAMQP(*eventBasicPublish, amqpRequest, basicMethodMap[40], connectionInfo, superTimer.CaptureTime, emitter)
			case *BasicDeliver:
				eventBasicDeliver.Body = f.Body
				superIdentifier.Protocol = &protocol
				emitAMQP(*eventBasicDeliver, amqpRequest, basicMethodMap[60], connectionInfo, superTimer.CaptureTime, emitter)
			default:
				body = nil
				frame = nil
			}

		case *MethodFrame:
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
				superIdentifier.Protocol = &protocol
				emitAMQP(*eventQueueBind, amqpRequest, queueMethodMap[20], connectionInfo, superTimer.CaptureTime, emitter)

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
				superIdentifier.Protocol = &protocol
				emitAMQP(*eventBasicConsume, amqpRequest, basicMethodMap[20], connectionInfo, superTimer.CaptureTime, emitter)

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
				superIdentifier.Protocol = &protocol
				emitAMQP(*eventQueueDeclare, amqpRequest, queueMethodMap[10], connectionInfo, superTimer.CaptureTime, emitter)

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
				superIdentifier.Protocol = &protocol
				emitAMQP(*eventExchangeDeclare, amqpRequest, exchangeMethodMap[10], connectionInfo, superTimer.CaptureTime, emitter)

			case *ConnectionStart:
				eventConnectionStart := &ConnectionStart{
					VersionMajor:     m.VersionMajor,
					VersionMinor:     m.VersionMinor,
					ServerProperties: m.ServerProperties,
					Mechanisms:       m.Mechanisms,
					Locales:          m.Locales,
				}
				superIdentifier.Protocol = &protocol
				emitAMQP(*eventConnectionStart, amqpRequest, connectionMethodMap[10], connectionInfo, superTimer.CaptureTime, emitter)

			case *ConnectionClose:
				eventConnectionClose := &ConnectionClose{
					ReplyCode: m.ReplyCode,
					ReplyText: m.ReplyText,
					ClassId:   m.ClassId,
					MethodId:  m.MethodId,
				}
				superIdentifier.Protocol = &protocol
				emitAMQP(*eventConnectionClose, amqpRequest, connectionMethodMap[50], connectionInfo, superTimer.CaptureTime, emitter)

			default:
				frame = nil

			}

		default:
			// log.Printf("unexpected frame: %+v\n", f)
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *api.MizuEntry {
	request := item.Pair.Request.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	service := "amqp"
	if resolvedDestination != "" {
		service = resolvedDestination
	} else if resolvedSource != "" {
		service = resolvedSource
	}

	summary := ""
	switch request["method"] {
	case basicMethodMap[40]:
		summary = reqDetails["Exchange"].(string)
		break
	case basicMethodMap[60]:
		summary = reqDetails["Exchange"].(string)
		break
	case exchangeMethodMap[10]:
		summary = reqDetails["Exchange"].(string)
		break
	case queueMethodMap[10]:
		summary = reqDetails["Queue"].(string)
		break
	case connectionMethodMap[10]:
		summary = fmt.Sprintf(
			"%s.%s",
			strconv.Itoa(int(reqDetails["VersionMajor"].(float64))),
			strconv.Itoa(int(reqDetails["VersionMinor"].(float64))),
		)
		break
	case connectionMethodMap[50]:
		summary = reqDetails["ReplyText"].(string)
		break
	case queueMethodMap[20]:
		summary = reqDetails["Queue"].(string)
		break
	case basicMethodMap[20]:
		summary = reqDetails["Queue"].(string)
		break
	}

	request["url"] = summary
	entryBytes, _ := json.Marshal(item.Pair)
	return &api.MizuEntry{
		Protocol: protocol,
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
		Outgoing:            item.ConnectionInfo.IsOutgoing,
		Request:             reqDetails,
		EntryId:             entryId,
		Entry:               string(entryBytes),
		Url:                 fmt.Sprintf("%s%s", service, summary),
		Method:              request["method"].(string),
		Status:              0,
		RequestSenderIp:     item.ConnectionInfo.ClientIP,
		Service:             service,
		Timestamp:           item.Timestamp,
		ElapsedTime:         0,
		Path:                summary,
		ResolvedSource:      resolvedSource,
		ResolvedDestination: resolvedDestination,
		SourceIp:            item.ConnectionInfo.ClientIP,
		DestinationIp:       item.ConnectionInfo.ServerIP,
		SourcePort:          item.ConnectionInfo.ClientPort,
		DestinationPort:     item.ConnectionInfo.ServerPort,
		IsOutgoing:          item.ConnectionInfo.IsOutgoing,
	}

}

func (d dissecting) Summarize(entry *api.MizuEntry) *api.BaseEntryDetails {
	return &api.BaseEntryDetails{
		Id:              entry.Id,
		Protocol:        protocol,
		Url:             entry.Url,
		RequestSenderIp: entry.RequestSenderIp,
		Service:         entry.Service,
		Summary:         entry.Path,
		StatusCode:      entry.Status,
		Method:          entry.Method,
		Timestamp:       entry.Timestamp,
		SourceIp:        entry.SourceIp,
		DestinationIp:   entry.DestinationIp,
		SourcePort:      entry.SourcePort,
		DestinationPort: entry.DestinationPort,
		IsOutgoing:      entry.IsOutgoing,
		Latency:         entry.ElapsedTime,
		Rules: api.ApplicableRules{
			Latency: 0,
			Status:  false,
		},
	}
}

func (d dissecting) Represent(entry *api.MizuEntry) (p api.Protocol, object []byte, bodySize int64, err error) {
	p = protocol
	bodySize = 0
	var root map[string]interface{}
	json.Unmarshal([]byte(entry.Entry), &root)
	representation := make(map[string]interface{}, 0)
	request := root["request"].(map[string]interface{})["payload"].(map[string]interface{})
	var repRequest []interface{}
	details := request["details"].(map[string]interface{})
	switch request["method"].(string) {
	case basicMethodMap[40]:
		repRequest = representBasicPublish(details)
		break
	case basicMethodMap[60]:
		repRequest = representBasicDeliver(details)
		break
	case queueMethodMap[10]:
		repRequest = representQueueDeclare(details)
		break
	case exchangeMethodMap[10]:
		repRequest = representExchangeDeclare(details)
		break
	case connectionMethodMap[10]:
		repRequest = representConnectionStart(details)
		break
	case connectionMethodMap[50]:
		repRequest = representConnectionClose(details)
		break
	case queueMethodMap[20]:
		repRequest = representQueueBind(details)
		break
	case basicMethodMap[20]:
		repRequest = representBasicConsume(details)
		break
	}
	representation["request"] = repRequest
	object, err = json.Marshal(representation)
	return
}

func (d dissecting) Macros() map[string]string {
	return map[string]string{
		`amqp`: fmt.Sprintf(`proto.abbr == "%s"`, protocol.Abbreviation),
	}
}

var Dissector dissecting
