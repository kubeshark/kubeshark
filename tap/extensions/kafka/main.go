package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"

	"github.com/up9inc/mizu/tap/api"
)

var _protocol api.Protocol = api.Protocol{
	Name:            "kafka",
	LongName:        "Apache Kafka Protocol",
	Abbreviation:    "KAFKA",
	BackgroundColor: "#000000",
	ForegroundColor: "#ffffff",
	FontSize:        11,
	ReferenceLink:   "https://kafka.apache.org/protocol",
	Ports:           []string{"9092"},
}

func init() {
	log.Println("Initializing Kafka extension...")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = _protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s\n", _protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, emitter api.Emitter) {
	for {
		if isClient {
			ReadRequest(b, tcpID)
		} else {
			ReadResponse(b, tcpID, emitter)
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *api.MizuEntry {
	request := item.Pair.Request.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	service := fmt.Sprintf("kafka")
	apiKey := ApiKey(reqDetails["ApiKey"].(float64))

	summary := ""
	switch apiKey {
	case Metadata:
		_topics := reqDetails["Payload"].(map[string]interface{})["Topics"]
		if _topics == nil {
			break
		}
		topics := _topics.([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["Name"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
		break
	case ApiVersions:
		summary = reqDetails["ClientID"].(string)
		break
	case Produce:
		topics := reqDetails["Payload"].(map[string]interface{})["TopicData"].([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["Topic"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
		break
	case Fetch:
		topics := reqDetails["Payload"].(map[string]interface{})["Topics"].([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["Topic"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
		break
	case ListOffsets:
		topics := reqDetails["Payload"].(map[string]interface{})["Topics"].([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["Name"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
		break
	case CreateTopics:
		topics := reqDetails["Payload"].(map[string]interface{})["Topics"].([]interface{})
		for _, topic := range topics {
			summary += fmt.Sprintf("%s, ", topic.(map[string]interface{})["Name"].(string))
		}
		if len(summary) > 0 {
			summary = summary[:len(summary)-2]
		}
		break
	case DeleteTopics:
		topicNames := reqDetails["TopicNames"].([]string)
		for _, name := range topicNames {
			summary += fmt.Sprintf("%s, ", name)
		}
		break
	}

	request["url"] = summary
	entryBytes, _ := json.Marshal(item.Pair)
	return &api.MizuEntry{
		ProtocolName:        _protocol.Name,
		EntryId:             entryId,
		Entry:               string(entryBytes),
		Url:                 fmt.Sprintf("%s%s", service, summary),
		Method:              apiNames[apiKey],
		Status:              0,
		RequestSenderIp:     item.ConnectionInfo.ClientIP,
		Service:             service,
		Timestamp:           item.Timestamp,
		Path:                summary,
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
		Id:              entry.EntryId,
		Protocol:        _protocol,
		Url:             entry.Url,
		RequestSenderIp: entry.RequestSenderIp,
		Service:         entry.Service,
		Summary:         entry.Path,
		StatusCode:      entry.Status,
		Method:          entry.Method,
		Timestamp:       entry.Timestamp,
		SourceIp:        entry.SourceIp,
		DestinationIp:   entry.DestinationIp,
		SourcePort:      entry.SourcePort,
		DestinationPort: entry.DestinationPort,
		IsOutgoing:      entry.IsOutgoing,
		Latency:         0,
		Rules: api.ApplicableRules{
			Latency: 0,
			Status:  false,
		},
	}
}

func (d dissecting) Represent(entry string) ([]byte, error) {
	var root map[string]interface{}
	json.Unmarshal([]byte(entry), &root)
	representation := make(map[string]interface{}, 0)
	request := root["request"].(map[string]interface{})["payload"].(map[string]interface{})
	response := root["response"].(map[string]interface{})["payload"].(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	resDetails := response["details"].(map[string]interface{})

	apiKey := ApiKey(reqDetails["ApiKey"].(float64))

	var repRequest []interface{}
	var repResponse []interface{}
	switch apiKey {
	case Metadata:
		repRequest = representMetadataRequest(reqDetails)
		repResponse = representMetadataResponse(resDetails)
		break
	case ApiVersions:
		repRequest = representApiVersionsRequest(reqDetails)
		repResponse = representApiVersionsResponse(resDetails)
	case Produce:
		repRequest = representProduceRequest(reqDetails)
		repResponse = representProduceResponse(resDetails)
		break
	}

	representation["request"] = repRequest
	representation["response"] = repResponse
	return json.Marshal(representation)
}

var Dissector dissecting
