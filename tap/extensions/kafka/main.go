package main

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
	extension.MatcherMap = reqResMatcher.openMessagesMap
}

func (d dissecting) Ping() {
	log.Printf("pong %s\n", _protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, superIdentifier *api.SuperIdentifier, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	for {
		if superIdentifier.Protocol != nil && superIdentifier.Protocol != &_protocol {
			return errors.New("Identified by another protocol")
		}

		if isClient {
			_, _, err := ReadRequest(b, tcpID, superTimer)
			if err != nil {
				return err
			}
			superIdentifier.Protocol = &_protocol
		} else {
			err := ReadResponse(b, tcpID, superTimer, emitter)
			if err != nil {
				return err
			}
			superIdentifier.Protocol = &_protocol
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *api.MizuEntry {
	request := item.Pair.Request.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	service := "kafka"
	if resolvedDestination != "" {
		service = resolvedDestination
	} else if resolvedSource != "" {
		service = resolvedSource
	}
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
		break
	case ApiVersions:
		summary = reqDetails["clientID"].(string)
		break
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
		break
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
		break
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
		break
	case CreateTopics:
		topics := reqDetails["payload"].(map[string]interface{})["topics"].([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["name"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
		break
	case DeleteTopics:
		topicNames := reqDetails["topicNames"].([]string)
		for _, name := range topicNames {
			summary += fmt.Sprintf("%s, ", name)
		}
		break
	}

	request["url"] = summary
	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	entryBytes, _ := json.Marshal(item.Pair)
	return &api.MizuEntry{
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
		Outgoing:            item.ConnectionInfo.IsOutgoing,
		Request:             reqDetails,
		Response:            item.Pair.Response.Payload.(map[string]interface{})["details"].(map[string]interface{}),
		EntryId:             entryId,
		Entry:               string(entryBytes),
		Url:                 fmt.Sprintf("%s%s", service, summary),
		Method:              apiNames[apiKey],
		Status:              0,
		RequestSenderIp:     item.ConnectionInfo.ClientIP,
		Service:             service,
		Timestamp:           item.Timestamp,
		StartTime:           item.Pair.Request.CaptureTime,
		ElapsedTime:         elapsedTime,
		Summary:             summary,
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
		Protocol:        _protocol,
		Url:             entry.Url,
		RequestSenderIp: entry.RequestSenderIp,
		Service:         entry.Service,
		Summary:         entry.Summary,
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

func (d dissecting) Represent(protoIn api.Protocol, request map[string]interface{}, response map[string]interface{}) (protoOut api.Protocol, object []byte, bodySize int64, err error) {
	protoOut = _protocol
	bodySize = 0
	representation := make(map[string]interface{}, 0)

	apiKey := ApiKey(request["apiKey"].(float64))

	var repRequest []interface{}
	var repResponse []interface{}
	switch apiKey {
	case Metadata:
		repRequest = representMetadataRequest(request)
		repResponse = representMetadataResponse(response)
		break
	case ApiVersions:
		repRequest = representApiVersionsRequest(request)
		repResponse = representApiVersionsResponse(response)
		break
	case Produce:
		repRequest = representProduceRequest(request)
		repResponse = representProduceResponse(response)
		break
	case Fetch:
		repRequest = representFetchRequest(request)
		repResponse = representFetchResponse(response)
		break
	case ListOffsets:
		repRequest = representListOffsetsRequest(request)
		repResponse = representListOffsetsResponse(response)
		break
	case CreateTopics:
		repRequest = representCreateTopicsRequest(request)
		repResponse = representCreateTopicsResponse(response)
		break
	case DeleteTopics:
		repRequest = representDeleteTopicsRequest(request)
		repResponse = representDeleteTopicsResponse(response)
		break
	}

	representation["request"] = repRequest
	representation["response"] = repResponse
	object, err = json.Marshal(representation)
	return
}

func (d dissecting) Macros() map[string]string {
	return map[string]string{
		`kafka`: fmt.Sprintf(`proto.abbr == "%s"`, _protocol.Abbreviation),
	}
}

var Dissector dissecting
