package tap

import (
	"fmt"
	"io"
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

func printEventBasicPublish(eventBasicPublish basicPublish) {
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

func printEventBasicDeliver(eventBasicDeliver basicDeliver) {
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

func printEventQueueDeclare(eventQueueDeclare queueDeclare) {
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

func printEventExchangeDeclare(eventExchangeDeclare exchangeDeclare) {
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

func printEventConnectionStart(eventConnectionStart connectionStart) {
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

func printEventConnectionClose(eventConnectionClose connectionClose) {
	fmt.Printf(
		"[%s] ReplyCode: %d, ReplyText: %s, ClassId: %d, MethodId: %d\n",
		connectionMethodMap[50],
		eventConnectionClose.ReplyCode,
		eventConnectionClose.ReplyText,
		eventConnectionClose.ClassId,
		eventConnectionClose.MethodId,
	)
}

func printEventQueueBind(eventQueueBind queueBind) {
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

func printEventBasicConsume(eventBasicConsume basicConsume) {
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

func (r *reader) Read() error {
	var remaining int
	var header *headerFrame
	var body []byte

	eventBasicPublish := &basicPublish{
		Exchange:   "",
		RoutingKey: "",
		Mandatory:  false,
		Immediate:  false,
		Body:       nil,
		Properties: properties{},
	}

	eventBasicDeliver := &basicDeliver{
		ConsumerTag: "",
		DeliveryTag: 0,
		Redelivered: false,
		Exchange:    "",
		RoutingKey:  "",
		Properties:  properties{},
		Body:        nil,
	}

	var lastMethodFrameMessage message

	for {
		frame, err := r.ReadFrame()
		if err == io.EOF {
			// We must read until we see an EOF... very important!
			return nil
		} else if err != nil {
			// log.Println("Error reading stream", h.net, h.transport, ":", err)
		}

		switch f := frame.(type) {
		case *heartbeatFrame:
			// drop

		case *headerFrame:
			// start content state
			header = f
			remaining = int(header.Size)
			switch lastMethodFrameMessage.(type) {
			case *basicPublish:
				eventBasicPublish.Properties = header.Properties
			case *basicDeliver:
				eventBasicDeliver.Properties = header.Properties
			default:
			}

		case *bodyFrame:
			// continue until terminated
			body = append(body, f.Body...)
			remaining -= len(f.Body)
			switch lastMethodFrameMessage.(type) {
			case *basicPublish:
				eventBasicPublish.Body = f.Body
				printEventBasicPublish(*eventBasicPublish)
			case *basicDeliver:
				eventBasicDeliver.Body = f.Body
				printEventBasicDeliver(*eventBasicDeliver)
			default:
			}

		case *methodFrame:
			lastMethodFrameMessage = f.Method
			switch m := f.Method.(type) {
			case *basicPublish:
				eventBasicPublish.Exchange = m.Exchange
				eventBasicPublish.RoutingKey = m.RoutingKey
				eventBasicPublish.Mandatory = m.Mandatory
				eventBasicPublish.Immediate = m.Immediate

			case *queueBind:
				eventQueueBind := &queueBind{
					Queue:      m.Queue,
					Exchange:   m.Exchange,
					RoutingKey: m.RoutingKey,
					NoWait:     m.NoWait,
					Arguments:  m.Arguments,
				}
				printEventQueueBind(*eventQueueBind)

			case *basicConsume:
				eventBasicConsume := &basicConsume{
					Queue:       m.Queue,
					ConsumerTag: m.ConsumerTag,
					NoLocal:     m.NoLocal,
					NoAck:       m.NoAck,
					Exclusive:   m.Exclusive,
					NoWait:      m.NoWait,
					Arguments:   m.Arguments,
				}
				printEventBasicConsume(*eventBasicConsume)

			case *basicDeliver:
				eventBasicDeliver.ConsumerTag = m.ConsumerTag
				eventBasicDeliver.DeliveryTag = m.DeliveryTag
				eventBasicDeliver.Redelivered = m.Redelivered
				eventBasicDeliver.Exchange = m.Exchange
				eventBasicDeliver.RoutingKey = m.RoutingKey

			case *queueDeclare:
				eventQueueDeclare := &queueDeclare{
					Queue:      m.Queue,
					Passive:    m.Passive,
					Durable:    m.Durable,
					AutoDelete: m.AutoDelete,
					Exclusive:  m.Exclusive,
					NoWait:     m.NoWait,
					Arguments:  m.Arguments,
				}
				printEventQueueDeclare(*eventQueueDeclare)

			case *exchangeDeclare:
				eventExchangeDeclare := &exchangeDeclare{
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

			case *connectionStart:
				eventConnectionStart := &connectionStart{
					VersionMajor:     m.VersionMajor,
					VersionMinor:     m.VersionMinor,
					ServerProperties: m.ServerProperties,
					Mechanisms:       m.Mechanisms,
					Locales:          m.Locales,
				}
				printEventConnectionStart(*eventConnectionStart)

			case *connectionClose:
				eventConnectionClose := &connectionClose{
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
