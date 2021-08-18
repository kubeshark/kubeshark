package api

import (
	"bufio"
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
	Type           string
	Timestamp      int64
	ConnectionInfo *ConnectionInfo
	Data           *RequestResponsePair
}

type Dissector interface {
	Register(*Extension)
	Ping()
	Dissect(b *bufio.Reader, isClient bool, tcpID *TcpID) *OutputChannelItem
}
