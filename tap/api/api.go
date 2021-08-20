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
	Orig        interface{}
}

type RequestResponsePair struct {
	Request  GenericMessage `json:"request"`
	Response GenericMessage `json:"response"`
}

type OutputChannelItem struct {
	Protocol       string
	Timestamp      int64
	ConnectionInfo *ConnectionInfo
	Data           *RequestResponsePair
}

type Dissector interface {
	Register(*Extension)
	Ping()
	Dissect(b *bufio.Reader, isClient bool, tcpID *TcpID, emitter Emitter)
}

type Emitting struct {
	OutputChannel chan *OutputChannelItem
}

type Emitter interface {
	Emit(item *OutputChannelItem)
}

func (e *Emitting) Emit(item *OutputChannelItem) {
	log.Printf("item: %+v\n", item)
	log.Printf("item.Data: %+v\n", item.Data)
	log.Printf("item.Data.Request.Orig: %v\n", item.Data.Request.Orig)
	log.Printf("item.Data.Response.Orig: %v\n", item.Data.Response.Orig)
	e.OutputChannel <- item
}
