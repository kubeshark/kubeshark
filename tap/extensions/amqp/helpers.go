package main

import (
	"fmt"
	"time"

	"github.com/up9inc/mizu/tap/api"
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

type AMQPWrapper struct {
	Method  string
	Details interface{}
}

func emitBasicPublish(eventBasicPublish BasicPublish, connectionInfo *api.ConnectionInfo, emitter api.Emitter) {
	request := &api.GenericMessage{
		IsRequest:   true,
		CaptureTime: time.Now(),
		Payload: AMQPPayload{
			Type: "basic_publish",
			Data: &AMQPWrapper{
				Method:  "Basic Publish",
				Details: eventBasicPublish,
			},
		},
	}
	item := &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      time.Now().UnixNano() / int64(time.Millisecond),
		ConnectionInfo: nil,
		Pair: &api.RequestResponsePair{
			Request:  *request,
			Response: api.GenericMessage{},
		},
	}
	emitter.Emit(item)
}

func printEventBasicPublish(eventBasicPublish BasicPublish) {
	return
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

func printEventBasicDeliver(eventBasicDeliver BasicDeliver) {
	return
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

func printEventQueueDeclare(eventQueueDeclare QueueDeclare) {
	return
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

func printEventExchangeDeclare(eventExchangeDeclare ExchangeDeclare) {
	return
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

func printEventConnectionStart(eventConnectionStart ConnectionStart) {
	return
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

func printEventConnectionClose(eventConnectionClose ConnectionClose) {
	return
	fmt.Printf(
		"[%s] ReplyCode: %d, ReplyText: %s, ClassId: %d, MethodId: %d\n",
		connectionMethodMap[50],
		eventConnectionClose.ReplyCode,
		eventConnectionClose.ReplyText,
		eventConnectionClose.ClassId,
		eventConnectionClose.MethodId,
	)
}

func printEventQueueBind(eventQueueBind QueueBind) {
	return
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

func printEventBasicConsume(eventBasicConsume BasicConsume) {
	return
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
