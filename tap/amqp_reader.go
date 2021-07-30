package tap

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/up9inc/mizu/amqp"
)

var connectionMethodMap = map[int]string{
	10: "connection start",
	11: "connection start-ok",
	20: "connection secure",
	21: "connection secure-ok",
	30: "connection tune",
	31: "connection tune-ok",
	40: "connection open",
	41: "connection open-ok",
	50: "connection close",
	51: "connection close-ok",
	60: "connection blocked",
	61: "connection unblocked",
}

var channelMethodMap = map[int]string{
	10: "channel open",
	11: "channel open-ok",
	20: "channel flow",
	21: "channel flow-ok",
	40: "channel close",
	41: "channel close-ok",
}

var exchangeMethodMap = map[int]string{
	10: "exchange declare",
	11: "exchange declare-ok",
	20: "exchange delete",
	21: "exchange delete-ok",
	30: "exchange bind",
	31: "exchange bind-ok",
	40: "exchange unbind",
	51: "exchange unbind-ok",
}

var queueMethodMap = map[int]string{
	10: "queue declare",
	11: "queue declare-ok",
	20: "queue bind",
	21: "queue bind-ok",
	50: "queue unbind",
	51: "queue unbind-ok",
	30: "queue purge",
	31: "queue purge-ok",
	40: "queue delete",
	41: "queue delete-ok",
}

var basicMethodMap = map[int]string{
	10:  "basic qos",
	11:  "basic qos-ok",
	20:  "basic consume",
	21:  "basic consume-ok",
	30:  "basic cancel",
	31:  "basic cancel-ok",
	40:  "basic publish",
	50:  "basic return",
	60:  "basic deliver",
	70:  "basic get",
	71:  "basic get-ok",
	72:  "basic get-empty",
	80:  "basic ack",
	90:  "basic reject",
	100: "basic recover-async",
	110: "basic recover",
	111: "basic recover-ok",
	120: "basic nack",
}

var txMethodMap = map[int]string{
	10: "tx select",
	11: "tx select-ok",
	20: "tx commit",
	21: "tx commit-ok",
	30: "tx rollback",
	31: "tx rollback-ok",
}

func printEventBasicPublish(eventBasicPublish amqp.BasicPublish) {
	fmt.Printf(
		"[%s] Exchange: %s, RoutingKey: %s, Mandatory: %t, Immediate: %t, Properties: %v, Body: %s\n",
		basicMethodMap[40],
		eventBasicPublish.Exchange,
		eventBasicPublish.RoutingKey,
		eventBasicPublish.Mandatory,
		eventBasicPublish.Immediate,
		eventBasicPublish.Properties,
		eventBasicPublish.Body,
	)
}

func printEventBasicDeliver(eventBasicDeliver amqp.BasicDeliver) {
	fmt.Printf(
		"[%s] ConsumerTag: %s, DeliveryTag: %d, Redelivered: %t, Exchange: %s, RoutingKey: %s, Properties: %v, Body: %s\n",
		basicMethodMap[60],
		eventBasicDeliver.ConsumerTag,
		eventBasicDeliver.DeliveryTag,
		eventBasicDeliver.Redelivered,
		eventBasicDeliver.Exchange,
		eventBasicDeliver.RoutingKey,
		eventBasicDeliver.Properties,
		eventBasicDeliver.Body,
	)
}

func printEventQueueDeclare(eventQueueDeclare amqp.QueueDeclare) {
	fmt.Printf(
		"[%s] Queue: %s, Passive: %t, Durable: %t, AutoDelete: %t, Exclusive: %t, NoWait: %t, Arguments: %v\n",
		queueMethodMap[10],
		eventQueueDeclare.Queue,
		eventQueueDeclare.Passive,
		eventQueueDeclare.Durable,
		eventQueueDeclare.AutoDelete,
		eventQueueDeclare.Exclusive,
		eventQueueDeclare.NoWait,
		eventQueueDeclare.Arguments,
	)
}

func printEventExchangeDeclare(eventExchangeDeclare amqp.ExchangeDeclare) {
	fmt.Printf(
		"[%s] Exchange: %s, Type: %s, Passive: %t, Durable: %t, AutoDelete: %t, Internal: %t, NoWait: %t, Arguments: %v\n",
		exchangeMethodMap[10],
		eventExchangeDeclare.Exchange,
		eventExchangeDeclare.Type,
		eventExchangeDeclare.Passive,
		eventExchangeDeclare.Durable,
		eventExchangeDeclare.AutoDelete,
		eventExchangeDeclare.Internal,
		eventExchangeDeclare.NoWait,
		eventExchangeDeclare.Arguments,
	)
}

func printEventConnectionStart(eventConnectionStart amqp.ConnectionStart) {
	fmt.Printf(
		"[%s] Version: %d.%d, ServerProperties: %v, Mechanisms: %s, Locales: %s\n",
		connectionMethodMap[10],
		eventConnectionStart.VersionMajor,
		eventConnectionStart.VersionMinor,
		eventConnectionStart.ServerProperties,
		eventConnectionStart.Mechanisms,
		eventConnectionStart.Locales,
	)
}

func printEventConnectionClose(eventConnectionClose amqp.ConnectionClose) {
	fmt.Printf(
		"[%s] ReplyCode: %d, ReplyText: %s, ClassId: %d, MethodId: %d\n",
		connectionMethodMap[50],
		eventConnectionClose.ReplyCode,
		eventConnectionClose.ReplyText,
		eventConnectionClose.ClassId,
		eventConnectionClose.MethodId,
	)
}

func printEventQueueBind(eventQueueBind amqp.QueueBind) {
	fmt.Printf(
		"[%s] Queue: %s, Exchange: %s, RoutingKey: %s, NoWait: %t, Arguments: %v\n",
		queueMethodMap[20],
		eventQueueBind.Queue,
		eventQueueBind.Exchange,
		eventQueueBind.RoutingKey,
		eventQueueBind.NoWait,
		eventQueueBind.Arguments,
	)
}

func printEventBasicConsume(eventBasicConsume amqp.BasicConsume) {
	fmt.Printf(
		"[%s] Queue: %s, ConsumerTag: %s, NoLocal: %t, NoAck: %t, Exclusive: %t, NoWait: %t, Arguments: %v\n",
		basicMethodMap[20],
		eventBasicConsume.Queue,
		eventBasicConsume.ConsumerTag,
		eventBasicConsume.NoLocal,
		eventBasicConsume.NoAck,
		eventBasicConsume.Exclusive,
		eventBasicConsume.NoWait,
		eventBasicConsume.Arguments,
	)
}

type amqpReaderIO struct {
	msgQueue    chan httpReaderDataMsg // Channel of captured reassembled tcp payload
	data        []byte
	captureTime time.Time
}

type amqpReader struct {
	r amqp.Reader
}

func (h *amqpReaderIO) run(wg *sync.WaitGroup) {
	defer wg.Done()
	b := bufio.NewReader(h)
	r := amqpReader{amqp.Reader{R: b}}
	r.Parse()
}

func (h *amqpReaderIO) Read(p []byte) (int, error) {
	var msg httpReaderDataMsg
	ok := true
	for ok && len(h.data) == 0 {
		msg, ok = <-h.msgQueue
		h.data = msg.bytes
		h.captureTime = msg.timestamp
	}
	if !ok || len(h.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, h.data)
	h.data = h.data[l:]
	return l, nil
}

func (r *amqpReader) Parse() error {
	fmt.Println("Parse is called")
	var remaining int
	var header *amqp.HeaderFrame
	var body []byte

	eventBasicPublish := &amqp.BasicPublish{
		Exchange:   "",
		RoutingKey: "",
		Mandatory:  false,
		Immediate:  false,
		Body:       nil,
		Properties: amqp.Properties{},
	}

	eventBasicDeliver := &amqp.BasicDeliver{
		ConsumerTag: "",
		DeliveryTag: 0,
		Redelivered: false,
		Exchange:    "",
		RoutingKey:  "",
		Properties:  amqp.Properties{},
		Body:        nil,
	}

	var lastMethodFrameMessage amqp.Message

	for {
		frame, err := r.r.ReadFrame()
		if err == io.EOF {
			// We must read until we see an EOF... very important!
			return nil
		} else if err != nil {
			// log.Println("Error reading stream", h.net, h.transport, ":", err)
		}

		switch f := frame.(type) {
		case *amqp.HeartbeatFrame:
			// drop

		case *amqp.HeaderFrame:
			// start content state
			header = f
			remaining = int(header.Size)
			switch lastMethodFrameMessage.(type) {
			case *amqp.BasicPublish:
				eventBasicPublish.Properties = header.Properties
			case *amqp.BasicDeliver:
				eventBasicDeliver.Properties = header.Properties
			default:
			}

		case *amqp.BodyFrame:
			// continue until terminated
			body = append(body, f.Body...)
			remaining -= len(f.Body)
			switch lastMethodFrameMessage.(type) {
			case *amqp.BasicPublish:
				eventBasicPublish.Body = f.Body
				printEventBasicPublish(*eventBasicPublish)
			case *amqp.BasicDeliver:
				eventBasicDeliver.Body = f.Body
				printEventBasicDeliver(*eventBasicDeliver)
			default:
			}

		case *amqp.MethodFrame:
			lastMethodFrameMessage = f.Method
			switch m := f.Method.(type) {
			case *amqp.BasicPublish:
				eventBasicPublish.Exchange = m.Exchange
				eventBasicPublish.RoutingKey = m.RoutingKey
				eventBasicPublish.Mandatory = m.Mandatory
				eventBasicPublish.Immediate = m.Immediate

			case *amqp.QueueBind:
				eventQueueBind := &amqp.QueueBind{
					Queue:      m.Queue,
					Exchange:   m.Exchange,
					RoutingKey: m.RoutingKey,
					NoWait:     m.NoWait,
					Arguments:  m.Arguments,
				}
				printEventQueueBind(*eventQueueBind)

			case *amqp.BasicConsume:
				eventBasicConsume := &amqp.BasicConsume{
					Queue:       m.Queue,
					ConsumerTag: m.ConsumerTag,
					NoLocal:     m.NoLocal,
					NoAck:       m.NoAck,
					Exclusive:   m.Exclusive,
					NoWait:      m.NoWait,
					Arguments:   m.Arguments,
				}
				printEventBasicConsume(*eventBasicConsume)

			case *amqp.BasicDeliver:
				eventBasicDeliver.ConsumerTag = m.ConsumerTag
				eventBasicDeliver.DeliveryTag = m.DeliveryTag
				eventBasicDeliver.Redelivered = m.Redelivered
				eventBasicDeliver.Exchange = m.Exchange
				eventBasicDeliver.RoutingKey = m.RoutingKey

			case *amqp.QueueDeclare:
				eventQueueDeclare := &amqp.QueueDeclare{
					Queue:      m.Queue,
					Passive:    m.Passive,
					Durable:    m.Durable,
					AutoDelete: m.AutoDelete,
					Exclusive:  m.Exclusive,
					NoWait:     m.NoWait,
					Arguments:  m.Arguments,
				}
				printEventQueueDeclare(*eventQueueDeclare)

			case *amqp.ExchangeDeclare:
				eventExchangeDeclare := &amqp.ExchangeDeclare{
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

			case *amqp.ConnectionStart:
				eventConnectionStart := &amqp.ConnectionStart{
					VersionMajor:     m.VersionMajor,
					VersionMinor:     m.VersionMinor,
					ServerProperties: m.ServerProperties,
					Mechanisms:       m.Mechanisms,
					Locales:          m.Locales,
				}
				printEventConnectionStart(*eventConnectionStart)

			case *amqp.ConnectionClose:
				eventConnectionClose := &amqp.ConnectionClose{
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
