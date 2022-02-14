package kafka

import (
	"bufio"
	"encoding/json"
	"errors"
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

func init() {
	log.Println("Initializing Kafka extension...")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &_protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s", _protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, superIdentifier *api.SuperIdentifier, emitter api.Emitter, options *api.TrafficFilteringOptions, _reqResMatcher api.RequestResponseMatcher) error {
	reqResMatcher := _reqResMatcher.(*requestResponseMatcher)
	for {
		if superIdentifier.Protocol != nil && superIdentifier.Protocol != &_protocol {
			return errors.New("Identified by another protocol")
		}

		if isClient {
			_, _, err := ReadRequest(b, tcpID, counterPair, superTimer, reqResMatcher)
			if err != nil {
				return err
			}
			superIdentifier.Protocol = &_protocol
		} else {
			err := ReadResponse(b, tcpID, counterPair, superTimer, emitter, reqResMatcher)
			if err != nil {
				return err
			}
			superIdentifier.Protocol = &_protocol
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, resolvedSource string, resolvedDestination string) *api.Entry {
	request := item.Pair.Request.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	apiKey := ApiKey(reqDetails["apiKey"].(float64))

	summary := ""
	switch apiKey {
	case Metadata:
		_topics := reqDetails["payload"].(map[string]interface{})["topics"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["name"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
	case ApiVersions:
		summary = reqDetails["clientID"].(string)
	case Produce:
		_topics := reqDetails["payload"].(map[string]interface{})["topicData"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["topic"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
	case Fetch:
		_topics := reqDetails["payload"].(map[string]interface{})["topics"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["topic"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
	case ListOffsets:
		_topics := reqDetails["payload"].(map[string]interface{})["topics"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["name"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
	case CreateTopics:
		_topics := reqDetails["payload"].(map[string]interface{})["topics"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["name"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
	case DeleteTopics:
		if reqDetails["topicNames"] == nil {
			break
		}
		topicNames := reqDetails["topicNames"].([]string)
		for _, name := range topicNames {
			summary += fmt.Sprintf("%s, ", name)
		}
	}

	request["url"] = summary
	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	if elapsedTime < 0 {
		elapsedTime = 0
	}
	return &api.Entry{
		Protocol: _protocol,
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
		Outgoing:    item.ConnectionInfo.IsOutgoing,
		Request:     reqDetails,
		Response:    item.Pair.Response.Payload.(map[string]interface{})["details"].(map[string]interface{}),
		Method:      apiNames[apiKey],
		Status:      0,
		Timestamp:   item.Timestamp,
		StartTime:   item.Pair.Request.CaptureTime,
		ElapsedTime: elapsedTime,
		Summary:     summary,
		IsOutgoing:  item.ConnectionInfo.IsOutgoing,
	}
}

func (d dissecting) Represent(request map[string]interface{}, response map[string]interface{}) (object []byte, bodySize int64, err error) {
	bodySize = 0
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
