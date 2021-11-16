package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

func (d dissecting) Analyze(item *api.OutputChannelItem, resolvedSource string, resolvedDestination string) *api.MizuEntry {
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
	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	if elapsedTime < 0 {
		elapsedTime = 0
	}
	return &api.MizuEntry{
		Protocol: protocol,
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
		Response:            resDetails,
		Url:                 fmt.Sprintf("%s%s", service, summary),
		Method:              method,
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
		Protocol:        protocol,
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
	protoOut = protocol
	bodySize = 0
	representation := make(map[string]interface{}, 0)
	repRequest := representGeneric(request, `request.`)
	repResponse := representGeneric(response, `response.`)
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
