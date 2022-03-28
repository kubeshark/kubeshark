package amqp

import (
	"encoding/json"
	"fmt"
	"sort"
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

// var channelMethodMap = map[int]string{
// 	10: "channel open",
// 	11: "channel open-ok",
// 	20: "channel flow",
// 	21: "channel flow-ok",
// 	40: "channel close",
// 	41: "channel close-ok",
// }

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

// var txMethodMap = map[int]string{
// 	10: "tx select",
// 	11: "tx select-ok",
// 	20: "tx commit",
// 	21: "tx commit-ok",
// 	30: "tx rollback",
// 	31: "tx rollback-ok",
// }

type AMQPWrapper struct {
	Method  string      `json:"method"`
	Url     string      `json:"url"`
	Details interface{} `json:"details"`
}

func emitAMQP(event interface{}, _type string, method string, connectionInfo *api.ConnectionInfo, captureTime time.Time, captureSize int, emitter api.Emitter, capture api.Capture) {
	request := &api.GenericMessage{
		IsRequest:   true,
		CaptureTime: captureTime,
		Payload: AMQPPayload{
			Data: &AMQPWrapper{
				Method:  method,
				Url:     "",
				Details: event,
			},
		},
	}
	item := &api.OutputChannelItem{
		Protocol:       protocol,
		Capture:        capture,
		Timestamp:      captureTime.UnixNano() / int64(time.Millisecond),
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

	if properties["contentType"] != nil {
		contentType = properties["contentType"].(string)
	}
	if properties["contentEncoding"] != nil {
		contentEncoding = properties["contentEncoding"].(string)
	}
	if properties["deliveryMode"] != nil {
		deliveryMode = fmt.Sprintf("%g", properties["deliveryMode"].(float64))
	}
	if properties["priority"] != nil {
		priority = fmt.Sprintf("%g", properties["priority"].(float64))
	}
	if properties["correlationId"] != nil {
		correlationId = properties["correlationId"].(string)
	}
	if properties["replyTo"] != nil {
		replyTo = properties["replyTo"].(string)
	}
	if properties["expiration"] != nil {
		expiration = properties["expiration"].(string)
	}
	if properties["messageId"] != nil {
		messageId = properties["messageId"].(string)
	}
	if properties["timestamp"] != nil {
		timestamp = properties["timestamp"].(string)
	}
	if properties["type"] != nil {
		_type = properties["type"].(string)
	}
	if properties["userId"] != nil {
		userId = properties["userId"].(string)
	}
	if properties["appId"] != nil {
		appId = properties["appId"].(string)
	}

	props, _ := json.Marshal([]api.TableData{
		{
			Name:     "Content Type",
			Value:    contentType,
			Selector: `request.properties.contentType`,
		},
		{
			Name:     "Content Encoding",
			Value:    contentEncoding,
			Selector: `request.properties.contentEncoding`,
		},
		{
			Name:     "Delivery Mode",
			Value:    deliveryMode,
			Selector: `request.properties.deliveryMode`,
		},
		{
			Name:     "Priority",
			Value:    priority,
			Selector: `request.properties.priority`,
		},
		{
			Name:     "Correlation ID",
			Value:    correlationId,
			Selector: `request.properties.correlationId`,
		},
		{
			Name:     "Reply To",
			Value:    replyTo,
			Selector: `request.properties.replyTo`,
		},
		{
			Name:     "Expiration",
			Value:    expiration,
			Selector: `request.properties.expiration`,
		},
		{
			Name:     "Message ID",
			Value:    messageId,
			Selector: `request.properties.messageId`,
		},
		{
			Name:     "Timestamp",
			Value:    timestamp,
			Selector: `request.properties.timestamp`,
		},
		{
			Name:     "Type",
			Value:    _type,
			Selector: `request.properties.type`,
		},
		{
			Name:     "User ID",
			Value:    userId,
			Selector: `request.properties.userId`,
		},
		{
			Name:     "App ID",
			Value:    appId,
			Selector: `request.properties.appId`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Properties",
		Data:  string(props),
	})

	return rep, contentType, contentEncoding
}

func representBasicPublish(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Exchange",
			Value:    event["exchange"].(string),
			Selector: `request.exchange`,
		},
		{
			Name:     "Routing Key",
			Value:    event["routingKey"].(string),
			Selector: `request.routingKey`,
		},
		{
			Name:     "Mandatory",
			Value:    strconv.FormatBool(event["mandatory"].(bool)),
			Selector: `request.mandatory`,
		},
		{
			Name:     "Immediate",
			Value:    strconv.FormatBool(event["immediate"].(bool)),
			Selector: `request.immediate`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	properties := event["properties"].(map[string]interface{})
	rep, contentType, _ := representProperties(properties, rep)

	if properties["headers"] != nil {
		headers := make([]api.TableData, 0)
		for name, value := range properties["headers"].(map[string]interface{}) {
			headers = append(headers, api.TableData{
				Name:     name,
				Value:    value.(string),
				Selector: fmt.Sprintf(`request.properties.headers["%s"]`, name),
			})
		}
		sort.Slice(headers, func(i, j int) bool {
			return headers[i].Name < headers[j].Name
		})
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, api.SectionData{
			Type:  api.TABLE,
			Title: "Headers",
			Data:  string(headersMarshaled),
		})
	}

	if event["body"] != nil {
		rep = append(rep, api.SectionData{
			Type:     api.BODY,
			Title:    "Body",
			Encoding: "base64",
			MimeType: contentType,
			Data:     event["body"].(string),
			Selector: `request.body`,
		})
	}

	return rep
}

func representBasicDeliver(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	consumerTag := ""
	deliveryTag := ""
	redelivered := ""

	if event["consumerTag"] != nil {
		consumerTag = event["consumerTag"].(string)
	}
	if event["deliveryTag"] != nil {
		deliveryTag = fmt.Sprintf("%g", event["deliveryTag"].(float64))
	}
	if event["redelivered"] != nil {
		redelivered = strconv.FormatBool(event["redelivered"].(bool))
	}

	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Consumer Tag",
			Value:    consumerTag,
			Selector: `request.consumerTag`,
		},
		{
			Name:     "Delivery Tag",
			Value:    deliveryTag,
			Selector: `request.deliveryTag`,
		},
		{
			Name:     "Redelivered",
			Value:    redelivered,
			Selector: `request.redelivered`,
		},
		{
			Name:     "Exchange",
			Value:    event["exchange"].(string),
			Selector: `request.exchange`,
		},
		{
			Name:     "Routing Key",
			Value:    event["routingKey"].(string),
			Selector: `request.routingKey`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	properties := event["properties"].(map[string]interface{})
	rep, contentType, _ := representProperties(properties, rep)

	if properties["headers"] != nil {
		headers := make([]api.TableData, 0)
		for name, value := range properties["headers"].(map[string]interface{}) {
			headers = append(headers, api.TableData{
				Name:     name,
				Value:    value,
				Selector: fmt.Sprintf(`request.properties.headers["%s"]`, name),
			})
		}
		sort.Slice(headers, func(i, j int) bool {
			return headers[i].Name < headers[j].Name
		})
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, api.SectionData{
			Type:  api.TABLE,
			Title: "Headers",
			Data:  string(headersMarshaled),
		})
	}

	if event["body"] != nil {
		rep = append(rep, api.SectionData{
			Type:     api.BODY,
			Title:    "Body",
			Encoding: "base64",
			MimeType: contentType,
			Data:     event["body"].(string),
			Selector: `request.body`,
		})
	}

	return rep
}

func representQueueDeclare(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Queue",
			Value:    event["queue"].(string),
			Selector: `request.queue`,
		},
		{
			Name:     "Passive",
			Value:    strconv.FormatBool(event["passive"].(bool)),
			Selector: `request.queue`,
		},
		{
			Name:     "Durable",
			Value:    strconv.FormatBool(event["durable"].(bool)),
			Selector: `request.durable`,
		},
		{
			Name:     "Exclusive",
			Value:    strconv.FormatBool(event["exclusive"].(bool)),
			Selector: `request.exclusive`,
		},
		{
			Name:     "Auto Delete",
			Value:    strconv.FormatBool(event["autoDelete"].(bool)),
			Selector: `request.autoDelete`,
		},
		{
			Name:     "NoWait",
			Value:    strconv.FormatBool(event["noWait"].(bool)),
			Selector: `request.noWait`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	if event["arguments"] != nil {
		headers := make([]api.TableData, 0)
		for name, value := range event["arguments"].(map[string]interface{}) {
			headers = append(headers, api.TableData{
				Name:     name,
				Value:    value.(string),
				Selector: fmt.Sprintf(`request.arguments["%s"]`, name),
			})
		}
		sort.Slice(headers, func(i, j int) bool {
			return headers[i].Name < headers[j].Name
		})
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, api.SectionData{
			Type:  api.TABLE,
			Title: "Arguments",
			Data:  string(headersMarshaled),
		})
	}

	return rep
}

func representExchangeDeclare(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Exchange",
			Value:    event["exchange"].(string),
			Selector: `request.exchange`,
		},
		{
			Name:     "Type",
			Value:    event["type"].(string),
			Selector: `request.type`,
		},
		{
			Name:     "Passive",
			Value:    strconv.FormatBool(event["passive"].(bool)),
			Selector: `request.passive`,
		},
		{
			Name:     "Durable",
			Value:    strconv.FormatBool(event["durable"].(bool)),
			Selector: `request.durable`,
		},
		{
			Name:     "Auto Delete",
			Value:    strconv.FormatBool(event["autoDelete"].(bool)),
			Selector: `request.autoDelete`,
		},
		{
			Name:     "Internal",
			Value:    strconv.FormatBool(event["internal"].(bool)),
			Selector: `request.internal`,
		},
		{
			Name:     "NoWait",
			Value:    strconv.FormatBool(event["noWait"].(bool)),
			Selector: `request.noWait`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	if event["arguments"] != nil {
		headers := make([]api.TableData, 0)
		for name, value := range event["arguments"].(map[string]interface{}) {
			headers = append(headers, api.TableData{
				Name:     name,
				Value:    value.(string),
				Selector: fmt.Sprintf(`request.arguments["%s"]`, name),
			})
		}
		sort.Slice(headers, func(i, j int) bool {
			return headers[i].Name < headers[j].Name
		})
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, api.SectionData{
			Type:  api.TABLE,
			Title: "Arguments",
			Data:  string(headersMarshaled),
		})
	}

	return rep
}

func representConnectionStart(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Version Major",
			Value:    fmt.Sprintf("%g", event["versionMajor"].(float64)),
			Selector: `request.versionMajor`,
		},
		{
			Name:     "Version Minor",
			Value:    fmt.Sprintf("%g", event["versionMinor"].(float64)),
			Selector: `request.versionMinor`,
		},
		{
			Name:     "Mechanisms",
			Value:    event["mechanisms"].(string),
			Selector: `request.mechanisms`,
		},
		{
			Name:     "Locales",
			Value:    event["locales"].(string),
			Selector: `request.locales`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	if event["serverProperties"] != nil {
		headers := make([]api.TableData, 0)
		for name, value := range event["serverProperties"].(map[string]interface{}) {
			var outcome string
			switch v := value.(type) {
			case string:
				outcome = v
			case map[string]interface{}:
				x, _ := json.Marshal(value)
				outcome = string(x)
			default:
				panic("Unknown data type for the server property!")
			}
			headers = append(headers, api.TableData{
				Name:     name,
				Value:    outcome,
				Selector: fmt.Sprintf(`request.serverProperties["%s"]`, name),
			})
		}
		sort.Slice(headers, func(i, j int) bool {
			return headers[i].Name < headers[j].Name
		})
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, api.SectionData{
			Type:  api.TABLE,
			Title: "Server Properties",
			Data:  string(headersMarshaled),
		})
	}

	return rep
}

func representConnectionClose(event map[string]interface{}) []interface{} {
	replyCode := ""

	if event["replyCode"] != nil {
		replyCode = fmt.Sprintf("%g", event["replyCode"].(float64))
	}

	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Reply Code",
			Value:    replyCode,
			Selector: `request.replyCode`,
		},
		{
			Name:     "Reply Text",
			Value:    event["replyText"].(string),
			Selector: `request.replyText`,
		},
		{
			Name:     "Class ID",
			Value:    fmt.Sprintf("%g", event["classId"].(float64)),
			Selector: `request.classId`,
		},
		{
			Name:     "Method ID",
			Value:    fmt.Sprintf("%g", event["methodId"].(float64)),
			Selector: `request.methodId`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	return rep
}

func representQueueBind(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Queue",
			Value:    event["queue"].(string),
			Selector: `request.queue`,
		},
		{
			Name:     "Exchange",
			Value:    event["exchange"].(string),
			Selector: `request.exchange`,
		},
		{
			Name:     "RoutingKey",
			Value:    event["routingKey"].(string),
			Selector: `request.routingKey`,
		},
		{
			Name:     "NoWait",
			Value:    strconv.FormatBool(event["noWait"].(bool)),
			Selector: `request.noWait`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	if event["arguments"] != nil {
		headers := make([]api.TableData, 0)
		for name, value := range event["arguments"].(map[string]interface{}) {
			headers = append(headers, api.TableData{
				Name:     name,
				Value:    value.(string),
				Selector: fmt.Sprintf(`request.arguments["%s"]`, name),
			})
		}
		sort.Slice(headers, func(i, j int) bool {
			return headers[i].Name < headers[j].Name
		})
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, api.SectionData{
			Type:  api.TABLE,
			Title: "Arguments",
			Data:  string(headersMarshaled),
		})
	}

	return rep
}

func representBasicConsume(event map[string]interface{}) []interface{} {
	rep := make([]interface{}, 0)

	details, _ := json.Marshal([]api.TableData{
		{
			Name:     "Queue",
			Value:    event["queue"].(string),
			Selector: `request.queue`,
		},
		{
			Name:     "Consumer Tag",
			Value:    event["consumerTag"].(string),
			Selector: `request.consumerTag`,
		},
		{
			Name:     "No Local",
			Value:    strconv.FormatBool(event["noLocal"].(bool)),
			Selector: `request.noLocal`,
		},
		{
			Name:     "No Ack",
			Value:    strconv.FormatBool(event["noAck"].(bool)),
			Selector: `request.noAck`,
		},
		{
			Name:     "Exclusive",
			Value:    strconv.FormatBool(event["exclusive"].(bool)),
			Selector: `request.exclusive`,
		},
		{
			Name:     "NoWait",
			Value:    strconv.FormatBool(event["noWait"].(bool)),
			Selector: `request.noWait`,
		},
	})
	rep = append(rep, api.SectionData{
		Type:  api.TABLE,
		Title: "Details",
		Data:  string(details),
	})

	if event["arguments"] != nil {
		headers := make([]api.TableData, 0)
		for name, value := range event["arguments"].(map[string]interface{}) {
			headers = append(headers, api.TableData{
				Name:     name,
				Value:    value.(string),
				Selector: fmt.Sprintf(`request.arguments["%s"]`, name),
			})
		}
		sort.Slice(headers, func(i, j int) bool {
			return headers[i].Name < headers[j].Name
		})
		headersMarshaled, _ := json.Marshal(headers)
		rep = append(rep, api.SectionData{
			Type:  api.TABLE,
			Title: "Arguments",
			Data:  string(headersMarshaled),
		})
	}

	return rep
}
