package redis

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

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s", protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, reader api.TcpReader, options *api.TrafficFilteringOptions) error {
	reqResMatcher := reader.GetReqResMatcher().(*requestResponseMatcher)
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

		if reader.GetIsClient() {
			err = handleClientStream(reader.GetReadProgress(), reader.GetParent().GetOrigin(), reader.GetTcpID(), reader.GetCounterPair(), reader.GetCaptureTime(), reader.GetEmitter(), redisPacket, reqResMatcher)
		} else {
			err = handleServerStream(reader.GetReadProgress(), reader.GetParent().GetOrigin(), reader.GetTcpID(), reader.GetCounterPair(), reader.GetCaptureTime(), reader.GetEmitter(), redisPacket, reqResMatcher)
		}

		if err != nil {
			return err
		}
	}
}

func (d dissecting) Analyze(item *api.OutputChannelItem, resolvedSource string, resolvedDestination string, namespace string) *api.Entry {
	request := item.Pair.Request.Payload.(map[string]interface{})
	response := item.Pair.Response.Payload.(map[string]interface{})
	reqDetails := request["details"].(map[string]interface{})
	resDetails := response["details"].(map[string]interface{})

	elapsedTime := item.Pair.Response.CaptureTime.Sub(item.Pair.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	if elapsedTime < 0 {
		elapsedTime = 0
	}
	return &api.Entry{
		Protocol: protocol,
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
		Response:     resDetails,
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

	method := ""
	methodQuery := ""
	if entry.Request["command"] != nil {
		method = entry.Request["command"].(string)
		methodQuery = fmt.Sprintf(`request.command == "%s"`, method)
	}

	summary := ""
	summaryQuery := ""
	if entry.Request["key"] != nil {
		summary = entry.Request["key"].(string)
		summaryQuery = fmt.Sprintf(`request.key == "%s"`, summary)
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
	repRequest := representGeneric(request, `request.`)
	repResponse := representGeneric(response, `response.`)
	representation["request"] = repRequest
	representation["response"] = repResponse
	object, err = json.Marshal(representation)
	return
}

func (d dissecting) Macros() map[string]string {
	return map[string]string{
		`redis`: fmt.Sprintf(`proto.name == "%s"`, protocol.Name),
	}
}

func (d dissecting) NewResponseRequestMatcher() api.RequestResponseMatcher {
	return createResponseRequestMatcher()
}

var Dissector dissecting

func NewDissector() api.Dissector {
	return Dissector
}
