package kafka

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

var _protocol api.Protocol = api.Protocol{
	Name:            "kafka",
	LongName:        "Apache Kafka Protocol",
	Abbreviation:    "KAFKA",
	Macro:           "kafka",
	Version:         "12",
	BackgroundColor: "#000000",
	ForegroundColor: "#ffffff",
	FontSize:        11,
	ReferenceLink:   "https://kafka.apache.org/protocol",
	Ports:           []string{"9092"},
	Priority:        2,
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &_protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s", _protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, reader api.TcpReader, options *api.TrafficFilteringOptions) error {
	reqResMatcher := reader.GetReqResMatcher().(*requestResponseMatcher)
	for {
		if reader.GetIsClient() {
			_, _, err := ReadRequest(b, reader.GetTcpID(), reader.GetCounterPair(), reader.GetCaptureTime(), reqResMatcher)
			if err != nil {
				return err
			}
			reader.GetParent().SetProtocol(&_protocol)
		} else {
			err := ReadResponse(b, reader.GetParent().GetOrigin(), reader.GetTcpID(), reader.GetCounterPair(), reader.GetCaptureTime(), reader.GetEmitter(), reqResMatcher)
			if err != nil {
				return err
			}
			reader.GetParent().SetProtocol(&_protocol)
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, resolvedSource string, resolvedDestination string, namespace string) *api.Entry {
	request := item.Pair.Request.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})

	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	if elapsedTime < 0 {
		elapsedTime = 0
	}
	return &api.Entry{
		Protocol: _protocol,
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
		Response:     item.Pair.Response.Payload.(map[string]interface{})["details"].(map[string]interface{}),
		RequestSize:  item.Pair.Request.CaptureSize,
		ResponseSize: item.Pair.Response.CaptureSize,
		Timestamp:    item.Timestamp,
		StartTime:    item.Pair.Request.CaptureTime,
		ElapsedTime:  elapsedTime,
	}
}

func (d dissecting) Summarize(entry *api.Entry) *api.BaseEntry {
	status := 0
	statusQuery := ""

	apiKey := ApiKey(entry.Request["apiKey"].(float64))
	method := entry.Request["apiKeyName"].(string)
	methodQuery := fmt.Sprintf(`request.apiKeyName == "%s"`, method)

	summary := ""
	summaryQuery := ""
	switch apiKey {
	case Metadata:
		_topics := entry.Request["payload"].(map[string]interface{})["topics"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for i, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["name"].(string))
			summaryQuery += fmt.Sprintf(`request.payload.topics[%d].name == "%s" and`, i, summary)
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
			summaryQuery = summaryQuery[:len(summaryQuery)-4]
		}
	case ApiVersions:
		summary = entry.Request["clientID"].(string)
		summaryQuery = fmt.Sprintf(`request.clientID == "%s"`, summary)
	case Produce:
		_topics := entry.Request["payload"].(map[string]interface{})["topicData"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for i, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["topic"].(string))
			summaryQuery += fmt.Sprintf(`request.payload.topicData[%d].topic == "%s" and`, i, summary)
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
			summaryQuery = summaryQuery[:len(summaryQuery)-4]
		}
	case Fetch:
		_topics := entry.Request["payload"].(map[string]interface{})["topics"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for i, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["topic"].(string))
			summaryQuery += fmt.Sprintf(`request.payload.topics[%d].topic == "%s" and`, i, summary)
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
			summaryQuery = summaryQuery[:len(summaryQuery)-4]
		}
	case ListOffsets:
		_topics := entry.Request["payload"].(map[string]interface{})["topics"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for i, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["name"].(string))
			summaryQuery += fmt.Sprintf(`request.payload.topics[%d].name == "%s" and`, i, summary)
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
			summaryQuery = summaryQuery[:len(summaryQuery)-4]
		}
	case CreateTopics:
		_topics := entry.Request["payload"].(map[string]interface{})["topics"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for i, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["name"].(string))
			summaryQuery += fmt.Sprintf(`request.payload.topics[%d].name == "%s" and`, i, summary)
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
			summaryQuery = summaryQuery[:len(summaryQuery)-4]
		}
	case DeleteTopics:
		if entry.Request["topicNames"] == nil {
			break
		}
		topicNames := entry.Request["topicNames"].([]string)
		for i, name := range topicNames {
			summary += fmt.Sprintf("%s, ", name)
			summaryQuery += fmt.Sprintf(`request.topicNames[%d] == "%s" and`, i, summary)
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
			summaryQuery = summaryQuery[:len(summaryQuery)-4]
		}
	}

	return &api.BaseEntry{
		Id:             entry.Id,
		Protocol:       entry.Protocol,
		Capture:        entry.Capture,
		Summary:        summary,
		SummaryQuery:   summaryQuery,
		Status:         status,
		StatusQuery:    statusQuery,
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

	apiKey := ApiKey(request["apiKey"].(float64))

	var repRequest []interface{}
	var repResponse []interface{}
	switch apiKey {
	case Metadata:
		repRequest = representMetadataRequest(request)
		repResponse = representMetadataResponse(response)
	case ApiVersions:
		repRequest = representApiVersionsRequest(request)
		repResponse = representApiVersionsResponse(response)
	case Produce:
		repRequest = representProduceRequest(request)
		repResponse = representProduceResponse(response)
	case Fetch:
		repRequest = representFetchRequest(request)
		repResponse = representFetchResponse(response)
	case ListOffsets:
		repRequest = representListOffsetsRequest(request)
		repResponse = representListOffsetsResponse(response)
	case CreateTopics:
		repRequest = representCreateTopicsRequest(request)
		repResponse = representCreateTopicsResponse(response)
	case DeleteTopics:
		repRequest = representDeleteTopicsRequest(request)
		repResponse = representDeleteTopicsResponse(response)
	}

	representation["request"] = repRequest
	representation["response"] = repResponse
	object, err = json.Marshal(representation)
	return
}

func (d dissecting) Macros() map[string]string {
	return map[string]string{
		`kafka`: fmt.Sprintf(`proto.name == "%s"`, _protocol.Name),
	}
}

func (d dissecting) NewResponseRequestMatcher() api.RequestResponseMatcher {
	return createResponseRequestMatcher()
}

var Dissector dissecting

func NewDissector() api.Dissector {
	return Dissector
}
