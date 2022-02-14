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

func init() {
	log.Println("Initializing Redis extension...")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = &protocol
	extension.MatcherMap = reqResMatcher.openMessagesMap
}

func (d dissecting) Ping() {
	log.Printf("pong %s", protocol.Name)
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
			err = handleClientStream(tcpID, counterPair, superTimer, emitter, redisPacket)
		} else {
			err = handleServerStream(tcpID, counterPair, superTimer, emitter, redisPacket)
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
	return &api.Entry{
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
		Namespace:   namespace,
		Outgoing:    item.ConnectionInfo.IsOutgoing,
		Request:     reqDetails,
		Response:    resDetails,
		Method:      method,
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

var Dissector dissecting

func NewDissector() api.Dissector {
	return Dissector
}
