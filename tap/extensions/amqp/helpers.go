package main

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	Method  string      `json:"method"`
	Url     string      `json:"url"`
	Details interface{} `json:"details"`
}

func emitBasicPublish(event BasicPublish, connectionInfo *api.ConnectionInfo, emitter api.Emitter) {
	request := &api.GenericMessage{
		IsRequest:   true,
		CaptureTime: time.Now(),
		Payload: AMQPPayload{
			Type: "basic_publish",
			Data: &AMQPWrapper{
				Method:  basicMethodMap[40],
				Url:     event.Exchange,
				Details: event,
			},
		},
	}
	item := &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      time.Now().UnixNano() / int64(time.Millisecond),
		ConnectionInfo: connectionInfo,
		Pair: &api.RequestResponsePair{
			Request:  *request,
			Response: api.GenericMessage{},
		},
	}
	emitter.Emit(item)
}

func representProperties(properties map[string]interface{}, rep []interface{}) ([]interface{}, string, string) {
	contentType := ""
	contentEncoding := ""
	deliveryMode := ""
	priority := ""
	correlationId := ""
	replyTo := ""
	expiration := ""
	messageId := ""
	timestamp := ""
	_type := ""
	userId := ""
	appId := ""

	if properties["ContentType"] != nil {
		contentType = properties["ContentType"].(string)
	}
	if properties["ContentEncoding"] != nil {
		contentEncoding = properties["ContentEncoding"].(string)
	}
	if properties["Delivery Mode"] != nil {
		deliveryMode = fmt.Sprintf("%g", properties["DeliveryMode"].(float64))
	}
	if properties["Priority"] != nil {
		priority = fmt.Sprintf("%g", properties["Priority"].(float64))
	}
	if properties["CorrelationId"] != nil {
		correlationId = properties["CorrelationId"].(string)
	}
	if properties["ReplyTo"] != nil {
		replyTo = properties["ReplyTo"].(string)
	}
	if properties["Expiration"] != nil {
		expiration = properties["Expiration"].(string)
	}
	if properties["MessageId"] != nil {
		messageId = properties["MessageId"].(string)
	}
	if properties["Timestamp"] != nil {
		timestamp = properties["Timestamp"].(string)
	}
	if properties["Type"] != nil {
		_type = properties["Type"].(string)
	}
	if properties["UserId"] != nil {
		userId = properties["UserId"].(string)
	}
	if properties["AppId"] != nil {
		appId = properties["AppId"].(string)
	}

	props, _ := json.Marshal([]map[string]string{
		{
			"name":  "Content Type",
			"value": contentType,
		},
		{
			"name":  "Content Encoding",
			"value": contentEncoding,
		},
		{
			"name":  "Delivery Mode",
			"value": deliveryMode,
		},
		{
			"name":  "Priority",
			"value": priority,
		},
		{
			"name":  "Correlation ID",
			"value": correlationId,
		},
		{
			"name":  "Reply To",
			"value": replyTo,
		},
		{
			"name":  "Expiration",
			"value": expiration,
		},
		{
			"name":  "Message ID",
			"value": messageId,
		},
		{
			"name":  "Timestamp",
			"value": timestamp,
		},
		{
			"name":  "Type",
			"value": _type,
		},
		{
			"name":  "User ID",
			"value": userId,
		},
		{
			"name":  "App ID",
			"value": appId,
		},
	})
	rep = append(rep, map[string]string{
		"type":  "table",
		"title": "Properties",
		"data":  string(props),
	})

	return rep, contentType, contentEncoding
}

func representBasicPublish(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]map[string]string{
		{
			"name":  "Exchange",
			"value": event["Exchange"].(string),
		},
		{
			"name":  "Routing Key",
			"value": event["RoutingKey"].(string),
		},
		{
			"name":  "Mandatory",
			"value": strconv.FormatBool(event["Mandatory"].(bool)),
		},
		{
			"name":  "Immediate",
			"value": strconv.FormatBool(event["Immediate"].(bool)),
		},
	})
	rep = append(rep, map[string]string{
		"type":  "table",
		"title": "Details",
		"data":  string(details),
	})

	properties := event["Properties"].(map[string]interface{})
	rep, contentType, contentEncoding := representProperties(properties, rep)

	if properties["Headers"] != nil {
		headers := make([]map[string]string, 0)
		for name, value := range properties["Headers"].(map[string]interface{}) {
			headers = append(headers, map[string]string{
				"name":  name,
				"value": value.(string),
			})
		}
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, map[string]string{
			"type":  "table",
			"title": "Headers",
			"data":  string(headersMarshaled),
		})
	}

	if event["Body"] != nil {
		rep = append(rep, map[string]string{
			"type":      "body",
			"title":     "Body",
			"encoding":  contentEncoding,
			"mime_type": contentType,
			"data":      event["Body"].(string),
		}) // FIXME: `Body` value seems wrong
	}

	return rep
}

func emitBasicDeliver(event BasicPublish, connectionInfo *api.ConnectionInfo, emitter api.Emitter) {
	request := &api.GenericMessage{
		IsRequest:   true,
		CaptureTime: time.Now(),
		Payload: AMQPPayload{
			Type: "basic_deliver",
			Data: &AMQPWrapper{
				Method:  basicMethodMap[60],
				Url:     event.Exchange,
				Details: event,
			},
		},
	}
	item := &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      time.Now().UnixNano() / int64(time.Millisecond),
		ConnectionInfo: connectionInfo,
		Pair: &api.RequestResponsePair{
			Request:  *request,
			Response: api.GenericMessage{},
		},
	}
	emitter.Emit(item)
}

func representBasicDeliver(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	consumerTag := ""
	deliveryTag := ""
	redelivered := ""

	if event["ConsumerTag"] != nil {
		consumerTag = event["ConsumerTag"].(string)
	}
	if event["DeliveryTag"] != nil {
		deliveryTag = fmt.Sprintf("%g", event["DeliveryTag"].(float64))
	}
	if event["Redelivered"] != nil {
		redelivered = strconv.FormatBool(event["Redelivered"].(bool))
	}

	details, _ := json.Marshal([]map[string]string{
		{
			"name":  "Consumer Tag",
			"value": consumerTag,
		},
		{
			"name":  "Delivery Tag",
			"value": deliveryTag,
		},
		{
			"name":  "Redelivered",
			"value": redelivered,
		},
		{
			"name":  "Exchange",
			"value": event["Exchange"].(string),
		},
		{
			"name":  "Routing Key",
			"value": event["RoutingKey"].(string),
		},
	})
	rep = append(rep, map[string]string{
		"type":  "table",
		"title": "Details",
		"data":  string(details),
	})

	properties := event["Properties"].(map[string]interface{})
	rep, contentType, contentEncoding := representProperties(properties, rep)

	if properties["Headers"] != nil {
		headers := make([]map[string]string, 0)
		for name, value := range properties["Headers"].(map[string]interface{}) {
			headers = append(headers, map[string]string{
				"name":  name,
				"value": value.(string),
			})
		}
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, map[string]string{
			"type":  "table",
			"title": "Headers",
			"data":  string(headersMarshaled),
		})
	}

	if event["Body"] != nil {
		rep = append(rep, map[string]string{
			"type":      "body",
			"title":     "Body",
			"encoding":  contentEncoding,
			"mime_type": contentType,
			"data":      event["Body"].(string),
		}) // FIXME: `Body` value seems wrong
	}

	return rep
}

func emitQueueDeclare(event QueueDeclare, connectionInfo *api.ConnectionInfo, emitter api.Emitter) {
	request := &api.GenericMessage{
		IsRequest:   true,
		CaptureTime: time.Now(),
		Payload: AMQPPayload{
			Type: "queue_declare",
			Data: &AMQPWrapper{
				Method:  queueMethodMap[10],
				Url:     event.Queue,
				Details: event,
			},
		},
	}
	item := &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      time.Now().UnixNano() / int64(time.Millisecond),
		ConnectionInfo: connectionInfo,
		Pair: &api.RequestResponsePair{
			Request:  *request,
			Response: api.GenericMessage{},
		},
	}
	emitter.Emit(item)
}

func representQueueDeclare(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]map[string]string{
		{
			"name":  "Queue",
			"value": event["Queue"].(string),
		},
		{
			"name":  "Passive",
			"value": strconv.FormatBool(event["Passive"].(bool)),
		},
		{
			"name":  "Durable",
			"value": strconv.FormatBool(event["Durable"].(bool)),
		},
		{
			"name":  "Exclusive",
			"value": strconv.FormatBool(event["Exclusive"].(bool)),
		},
		{
			"name":  "Auto Delete",
			"value": strconv.FormatBool(event["AutoDelete"].(bool)),
		},
		{
			"name":  "NoWait",
			"value": strconv.FormatBool(event["NoWait"].(bool)),
		},
	})
	rep = append(rep, map[string]string{
		"type":  "table",
		"title": "Details",
		"data":  string(details),
	})

	if event["Arguments"] != nil {
		headers := make([]map[string]string, 0)
		for name, value := range event["Arguments"].(map[string]interface{}) {
			headers = append(headers, map[string]string{
				"name":  name,
				"value": value.(string),
			})
		}
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, map[string]string{
			"type":  "table",
			"title": "Arguments",
			"data":  string(headersMarshaled),
		})
	}

	return rep
}

func emitExchangeDeclare(event ExchangeDeclare, connectionInfo *api.ConnectionInfo, emitter api.Emitter) {
	request := &api.GenericMessage{
		IsRequest:   true,
		CaptureTime: time.Now(),
		Payload: AMQPPayload{
			Type: "exchange_declare",
			Data: &AMQPWrapper{
				Method:  exchangeMethodMap[10],
				Url:     event.Exchange,
				Details: event,
			},
		},
	}
	item := &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      time.Now().UnixNano() / int64(time.Millisecond),
		ConnectionInfo: connectionInfo,
		Pair: &api.RequestResponsePair{
			Request:  *request,
			Response: api.GenericMessage{},
		},
	}
	emitter.Emit(item)
}

func representExchangeDeclare(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]map[string]string{
		{
			"name":  "Exchange",
			"value": event["Exchange"].(string),
		},
		{
			"name":  "Type",
			"value": event["Type"].(string),
		},
		{
			"name":  "Passive",
			"value": strconv.FormatBool(event["Passive"].(bool)),
		},
		{
			"name":  "Durable",
			"value": strconv.FormatBool(event["Durable"].(bool)),
		},
		{
			"name":  "Auto Delete",
			"value": strconv.FormatBool(event["AutoDelete"].(bool)),
		},
		{
			"name":  "Internal",
			"value": strconv.FormatBool(event["Internal"].(bool)),
		},
		{
			"name":  "NoWait",
			"value": strconv.FormatBool(event["NoWait"].(bool)),
		},
	})
	rep = append(rep, map[string]string{
		"type":  "table",
		"title": "Details",
		"data":  string(details),
	})

	if event["Arguments"] != nil {
		headers := make([]map[string]string, 0)
		for name, value := range event["Arguments"].(map[string]interface{}) {
			headers = append(headers, map[string]string{
				"name":  name,
				"value": value.(string),
			})
		}
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, map[string]string{
			"type":  "table",
			"title": "Arguments",
			"data":  string(headersMarshaled),
		})
	}

	return rep
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
