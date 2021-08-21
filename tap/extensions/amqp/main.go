package main

import (
	"bufio"
	"io"
	"log"

	"github.com/up9inc/mizu/tap/api"
)

var protocol api.Protocol = api.Protocol{
	Name:            "amqp",
	LongName:        "Advanced Message Queuing Protocol",
	Abbreviation:    "AMQP",
	BackgroundColor: "#ff6600",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://www.rabbitmq.com/amqp-0-9-1-reference.html",
	Ports:           []string{"5671", "5672"},
}

func init() {
	log.Println("Initializing AMQP extension...")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s\n", protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, emitter api.Emitter) {
	r := AmqpReader{b}

	var remaining int
	var header *HeaderFrame
	var body []byte

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
			return
		} else if err != nil {
			// log.Println("Error reading stream", h.net, h.transport, ":", err)
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
			}

		case *BodyFrame:
			// continue until terminated
			body = append(body, f.Body...)
			remaining -= len(f.Body)
			switch lastMethodFrameMessage.(type) {
			case *BasicPublish:
				eventBasicPublish.Body = f.Body
				printEventBasicPublish(*eventBasicPublish)
			case *BasicDeliver:
				eventBasicDeliver.Body = f.Body
				printEventBasicDeliver(*eventBasicDeliver)
			default:
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
				printEventQueueBind(*eventQueueBind)

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
				printEventBasicConsume(*eventBasicConsume)

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
				printEventQueueDeclare(*eventQueueDeclare)

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
				printEventExchangeDeclare(*eventExchangeDeclare)

			case *ConnectionStart:
				eventConnectionStart := &ConnectionStart{
					VersionMajor:     m.VersionMajor,
					VersionMinor:     m.VersionMinor,
					ServerProperties: m.ServerProperties,
					Mechanisms:       m.Mechanisms,
					Locales:          m.Locales,
				}
				printEventConnectionStart(*eventConnectionStart)

			case *ConnectionClose:
				eventConnectionClose := &ConnectionClose{
					ReplyCode: m.ReplyCode,
					ReplyText: m.ReplyText,
					ClassId:   m.ClassId,
					MethodId:  m.MethodId,
				}
				printEventConnectionClose(*eventConnectionClose)

			default:

			}

		default:
			// fmt.Printf("unexpected frame: %+v\n", f)
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *api.MizuEntry {
	// TODO: Implement
	return nil
}

func (d dissecting) Summarize(entry *api.MizuEntry) *api.BaseEntryDetails {
	// TODO: Implement
	return nil
}

func (d dissecting) Represent(entry string) ([]byte, error) {
	// TODO: Implement
	return nil, nil
}

var Dissector dissecting
