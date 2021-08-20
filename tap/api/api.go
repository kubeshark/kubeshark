package api

import (
	"bufio"
	"log"
	"plugin"
	"time"
)

type Extension struct {
	Name          string
	Path          string
	Plug          *plugin.Plugin
	InboundPorts  []string
	OutboundPorts []string
	Dissector     Dissector
}

type ConnectionInfo struct {
	ClientIP   string
	ClientPort string
	ServerIP   string
	ServerPort string
	IsOutgoing bool
}

type TcpID struct {
	SrcIP   string
	DstIP   string
	SrcPort string
	DstPort string
	Ident   string
}

type GenericMessage struct {
	IsRequest   bool
	CaptureTime time.Time
	Payload     interface{}
}

type RequestResponsePair struct {
	Request  GenericMessage `json:"request"`
	Response GenericMessage `json:"response"`
}

type OutputChannelItem struct {
	Protocol       string
	Timestamp      int64
	ConnectionInfo *ConnectionInfo
	Pair           *RequestResponsePair
}

type Dissector interface {
	Register(*Extension)
	Ping()
	Dissect(b *bufio.Reader, isClient bool, tcpID *TcpID, emitter Emitter)
	Summarize(item *OutputChannelItem) *BaseEntryDetails
}

type Emitting struct {
	OutputChannel chan *OutputChannelItem
}

type Emitter interface {
	Emit(item *OutputChannelItem)
}

func (e *Emitting) Emit(item *OutputChannelItem) {
	log.Printf("item: %+v\n", item)
	log.Printf("item.Pair: %+v\n", item.Pair)
	log.Printf("item.Pair.Request.Payload: %v\n", item.Pair.Request.Payload)
	log.Printf("item.Pair.Response.Payload: %v\n", item.Pair.Response.Payload)
	e.OutputChannel <- item
}

type BaseEntryDetails struct {
	Id              string          `json:"id,omitempty"`
	Url             string          `json:"url,omitempty"`
	RequestSenderIp string          `json:"requestSenderIp,omitempty"`
	Service         string          `json:"service,omitempty"`
	Path            string          `json:"path,omitempty"`
	StatusCode      int             `json:"statusCode,omitempty"`
	Method          string          `json:"method,omitempty"`
	Timestamp       int64           `json:"timestamp,omitempty"`
	IsOutgoing      bool            `json:"isOutgoing,omitempty"`
	Latency         int64           `json:"latency,omitempty"`
	Rules           ApplicableRules `json:"rules,omitempty"`
}

type ApplicableRules struct {
	Latency int64 `json:"latency,omitempty"`
	Status  bool  `json:"status,omitempty"`
}
