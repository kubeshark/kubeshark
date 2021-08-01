package tap

import (
	"bufio"
	"io"

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

// func printEventBasicPublish(eventBasicPublish *amqp.BasicPublish) {
// 	fmt.Printf(
// 		"Exchange: %s, RoutingKey: %s, Mandatory: %t, Immediate: %t, Properties: %v, Body: %s\n",
// 		eventBasicPublish.Exchange,
// 		eventBasicPublish.RoutingKey,
// 		eventBasicPublish.Mandatory,
// 		eventBasicPublish.Immediate,
// 		eventBasicPublish.Properties,
// 		eventBasicPublish.Body,
// 	)
// }

// func printEventBasicDeliver(eventBasicDeliver *amqp.BasicDeliver) {
// 	fmt.Printf(
// 		"ConsumerTag: %s, DeliveryTag: %d, Redelivered: %t, Exchange: %s, RoutingKey: %s, Properties: %v, Body: %s\n",
// 		eventBasicDeliver.ConsumerTag,
// 		eventBasicDeliver.DeliveryTag,
// 		eventBasicDeliver.Redelivered,
// 		eventBasicDeliver.Exchange,
// 		eventBasicDeliver.RoutingKey,
// 		eventBasicDeliver.Properties,
// 		eventBasicDeliver.Body,
// 	)
// }

// func printEventQueueDeclare(eventQueueDeclare *amqp.QueueDeclare) {
// 	fmt.Printf(
// 		"Queue: %s, Passive: %t, Durable: %t, AutoDelete: %t, Exclusive: %t, NoWait: %t, Arguments: %v\n",
// 		eventQueueDeclare.Queue,
// 		eventQueueDeclare.Passive,
// 		eventQueueDeclare.Durable,
// 		eventQueueDeclare.AutoDelete,
// 		eventQueueDeclare.Exclusive,
// 		eventQueueDeclare.NoWait,
// 		eventQueueDeclare.Arguments,
// 	)
// }

// func printEventExchangeDeclare(eventExchangeDeclare *amqp.ExchangeDeclare) {
// 	fmt.Printf(
// 		"Exchange: %s, Type: %s, Passive: %t, Durable: %t, AutoDelete: %t, Internal: %t, NoWait: %t, Arguments: %v\n",
// 		eventExchangeDeclare.Exchange,
// 		eventExchangeDeclare.Type,
// 		eventExchangeDeclare.Passive,
// 		eventExchangeDeclare.Durable,
// 		eventExchangeDeclare.AutoDelete,
// 		eventExchangeDeclare.Internal,
// 		eventExchangeDeclare.NoWait,
// 		eventExchangeDeclare.Arguments,
// 	)
// }

// func printEventConnectionStart(eventConnectionStart *amqp.ConnectionStart) {
// 	fmt.Printf(
// 		"Version: %d.%d, ServerProperties: %v, Mechanisms: %s, Locales: %s\n",
// 		eventConnectionStart.VersionMajor,
// 		eventConnectionStart.VersionMinor,
// 		eventConnectionStart.ServerProperties,
// 		eventConnectionStart.Mechanisms,
// 		eventConnectionStart.Locales,
// 	)
// }

// func printEventConnectionClose(eventConnectionClose *amqp.ConnectionClose) {
// 	fmt.Printf(
// 		"ReplyCode: %d, ReplyText: %s, ClassId: %d, MethodId: %d\n",
// 		eventConnectionClose.ReplyCode,
// 		eventConnectionClose.ReplyText,
// 		eventConnectionClose.ClassId,
// 		eventConnectionClose.MethodId,
// 	)
// }

// func printEventQueueBind(eventQueueBind *amqp.QueueBind) {
// 	fmt.Printf(
// 		"Queue: %s, Exchange: %s, RoutingKey: %s, NoWait: %t, Arguments: %v\n",
// 		eventQueueBind.Queue,
// 		eventQueueBind.Exchange,
// 		eventQueueBind.RoutingKey,
// 		eventQueueBind.NoWait,
// 		eventQueueBind.Arguments,
// 	)
// }

// func printEventBasicConsume(eventBasicConsume *amqp.BasicConsume) {
// 	fmt.Printf(
// 		"Queue: %s, ConsumerTag: %s, NoLocal: %t, NoAck: %t, Exclusive: %t, NoWait: %t, Arguments: %v\n",
// 		eventBasicConsume.Queue,
// 		eventBasicConsume.ConsumerTag,
// 		eventBasicConsume.NoLocal,
// 		eventBasicConsume.NoAck,
// 		eventBasicConsume.Exclusive,
// 		eventBasicConsume.NoWait,
// 		eventBasicConsume.Arguments,
// 	)
// }

type EventAMQP struct {
	Type string
	// Method          amqp.Message
	ConnectionStart *amqp.ConnectionStart
	ConnectionClose *amqp.ConnectionClose
	QueueDeclare    *amqp.QueueDeclare
	QueueBind       *amqp.QueueBind
	ExchangeDeclare *amqp.ExchangeDeclare
	BasicPublish    *amqp.BasicPublish
	BasicDeliver    *amqp.BasicDeliver
	BasicConsume    *amqp.BasicConsume
}

// func (e *EventAMQP) Print() {
// 	fmt.Printf("[%s] ", e.Type)
// 	switch e.Method.(type) {
// 	case *amqp.ConnectionStart:
// 		printEventConnectionStart(e.ConnectionStart)
// 	case *amqp.ConnectionClose:
// 		printEventConnectionClose(e.ConnectionClose)
// 	case *amqp.QueueDeclare:
// 		printEventQueueDeclare(e.QueueDeclare)
// 	case *amqp.QueueBind:
// 		printEventQueueBind(e.QueueBind)
// 	case *amqp.ExchangeDeclare:
// 		printEventExchangeDeclare(e.ExchangeDeclare)
// 	case *amqp.BasicPublish:
// 		printEventBasicPublish(e.BasicPublish)
// 	case *amqp.BasicDeliver:
// 		printEventBasicDeliver(e.BasicDeliver)
// 	case *amqp.BasicConsume:
// 		printEventBasicConsume(e.BasicConsume)
// 	}
// }

func (e *EventAMQP) Write(harWriter *HarWriter, connectionInfo *ConnectionInfo) {
	if harWriter != nil {
		harWriter.WriteAMQP(e, connectionInfo)
	}
}

type amqpReader struct {
	r          amqp.Reader
	ident      string
	tcpID      tcpID
	parent     *tcpStream
	isClient   bool
	isOutgoing bool
	harWriter  *HarWriter
}

func (h *amqpReader) run(b *bufio.Reader) {
	h.r = amqp.Reader{R: b}
	h.Parse()
}

func (h *amqpReader) Parse() error {
	var remaining int
	var header *amqp.HeaderFrame
	var body []byte

	eventAMQP := &EventAMQP{}
	connectionInfo := &ConnectionInfo{
		ClientIP:   h.tcpID.srcIP,
		ClientPort: h.tcpID.srcPort,
		ServerIP:   h.tcpID.dstIP,
		ServerPort: h.tcpID.dstPort,
		IsOutgoing: h.isOutgoing,
	}

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
		frame, err := h.r.ReadFrame()
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
				// eventAMQP.Print()
				eventAMQP.Write(h.harWriter, connectionInfo)
			case *amqp.BasicDeliver:
				eventBasicDeliver.Body = f.Body
				// eventAMQP.Print()
				eventAMQP.Write(h.harWriter, connectionInfo)
			default:
			}

		case *amqp.MethodFrame:
			lastMethodFrameMessage = f.Method
			// eventAMQP.Method = f.Method
			switch m := f.Method.(type) {
			case *amqp.BasicPublish:
				eventBasicPublish.Exchange = m.Exchange
				eventBasicPublish.RoutingKey = m.RoutingKey
				eventBasicPublish.Mandatory = m.Mandatory
				eventBasicPublish.Immediate = m.Immediate
				eventAMQP.Type = basicMethodMap[40]
				eventAMQP.BasicPublish = eventBasicPublish

			case *amqp.QueueBind:
				eventQueueBind := &amqp.QueueBind{
					Queue:      m.Queue,
					Exchange:   m.Exchange,
					RoutingKey: m.RoutingKey,
					NoWait:     m.NoWait,
					Arguments:  m.Arguments,
				}
				eventAMQP.Type = queueMethodMap[20]
				eventAMQP.QueueBind = eventQueueBind
				// eventAMQP.Print()
				eventAMQP.Write(h.harWriter, connectionInfo)

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
				eventAMQP.Type = basicMethodMap[20]
				eventAMQP.BasicConsume = eventBasicConsume
				// eventAMQP.Print()
				eventAMQP.Write(h.harWriter, connectionInfo)

			case *amqp.BasicDeliver:
				eventBasicDeliver.ConsumerTag = m.ConsumerTag
				eventBasicDeliver.DeliveryTag = m.DeliveryTag
				eventBasicDeliver.Redelivered = m.Redelivered
				eventBasicDeliver.Exchange = m.Exchange
				eventBasicDeliver.RoutingKey = m.RoutingKey
				eventAMQP.Type = basicMethodMap[60]
				eventAMQP.BasicDeliver = eventBasicDeliver

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
				eventAMQP.Type = queueMethodMap[10]
				eventAMQP.QueueDeclare = eventQueueDeclare
				// eventAMQP.Print()
				eventAMQP.Write(h.harWriter, connectionInfo)

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
				eventAMQP.Type = exchangeMethodMap[10]
				eventAMQP.ExchangeDeclare = eventExchangeDeclare
				// eventAMQP.Print()
				eventAMQP.Write(h.harWriter, connectionInfo)

			case *amqp.ConnectionStart:
				eventConnectionStart := &amqp.ConnectionStart{
					VersionMajor:     m.VersionMajor,
					VersionMinor:     m.VersionMinor,
					ServerProperties: m.ServerProperties,
					Mechanisms:       m.Mechanisms,
					Locales:          m.Locales,
				}
				eventAMQP.Type = connectionMethodMap[10]
				eventAMQP.ConnectionStart = eventConnectionStart
				// eventAMQP.Print()
				eventAMQP.Write(h.harWriter, connectionInfo)

			case *amqp.ConnectionClose:
				eventConnectionClose := &amqp.ConnectionClose{
					ReplyCode: m.ReplyCode,
					ReplyText: m.ReplyText,
					ClassId:   m.ClassId,
					MethodId:  m.MethodId,
				}
				eventAMQP.Type = connectionMethodMap[50]
				eventAMQP.ConnectionClose = eventConnectionClose
				// eventAMQP.Print()
				eventAMQP.Write(h.harWriter, connectionInfo)

			default:

			}

		default:
			// fmt.Printf("unexpected frame: %+v\n", f)
		}
	}
}
