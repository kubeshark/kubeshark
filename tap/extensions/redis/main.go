package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"

	"github.com/up9inc/mizu/tap/api"
)

var protocol api.Protocol = api.Protocol{
	Name:            "redis",
	LongName:        "Redis Serialization Protocol",
	Abbreviation:    "REDIS",
	Macro:           "redis",
	Version:         "3.x",
	BackgroundColor: "#a41e11",
	ForegroundColor: "#ffffff",
	FontSize:        11,
	ReferenceLink:   "https://redis.io/topics/protocol",
	Ports:           []string{"6379"},
	Priority:        3,
}

func init() {
	log.Println("Initializing Redis extension...")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &protocol
	extension.MatcherMap = reqResMatcher.openMessagesMap
}

func (d dissecting) Ping() {
	log.Printf("pong %s\n", protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, superIdentifier *api.SuperIdentifier, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	is := &RedisInputStream{
		Reader: b,
		Buf:    make([]byte, 8192),
	}
	proto := NewProtocol(is)
	for {
		redisPacket, err := proto.Read()
		if err != nil {
			return err
		}

		if isClient {
			handleClientStream(tcpID, counterPair, superTimer, emitter, redisPacket)
		} else {
			handleServerStream(tcpID, counterPair, superTimer, emitter, redisPacket)
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *api.MizuEntry {
	request := item.Pair.Request.Payload.(map[string]interface{})
	response := item.Pair.Response.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	resDetails := response["details"].(map[string]interface{})

	service := "redis"
	if resolvedDestination != "" {
		service = resolvedDestination
	} else if resolvedSource != "" {
		service = resolvedSource
	}

	method := ""
	if reqDetails["command"] != nil {
		method = reqDetails["command"].(string)
	}

	summary := ""
	if reqDetails["key"] != nil {
		summary = reqDetails["key"].(string)
	}

	request["url"] = summary
	entryBytes, _ := json.Marshal(item.Pair)
	return &api.MizuEntry{
		Protocol:            protocol,
		Request:             reqDetails,
		Response:            resDetails,
		EntryId:             entryId,
		Entry:               string(entryBytes),
		Url:                 fmt.Sprintf("%s%s", service, summary),
		Method:              method,
		Status:              0,
		RequestSenderIp:     item.ConnectionInfo.ClientIP,
		Service:             service,
		Timestamp:           item.Timestamp,
		ElapsedTime:         0,
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
		Id:              entry.Id,
		Protocol:        protocol,
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
		Latency:         entry.ElapsedTime,
		Rules: api.ApplicableRules{
			Latency: 0,
			Status:  false,
		},
	}
}

func (d dissecting) Represent(entry *api.MizuEntry) (p api.Protocol, object []byte, bodySize int64, err error) {
	p = protocol
	bodySize = 0
	var root map[string]interface{}
	json.Unmarshal([]byte(entry.Entry), &root)
	representation := make(map[string]interface{}, 0)
	request := root["request"].(map[string]interface{})["payload"].(map[string]interface{})
	response := root["response"].(map[string]interface{})["payload"].(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	resDetails := response["details"].(map[string]interface{})
	repRequest := representGeneric(reqDetails)
	repResponse := representGeneric(resDetails)
	representation["request"] = repRequest
	representation["response"] = repResponse
	object, err = json.Marshal(representation)
	return
}

func (d dissecting) Macros() map[string]string {
	return map[string]string{
		`redis`: fmt.Sprintf(`proto.abbr == "%s"`, protocol.Abbreviation),
	}
}

var Dissector dissecting
