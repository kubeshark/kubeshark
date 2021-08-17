package api

import (
	"bufio"
	"plugin"
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

type Dissector interface {
	Register(*Extension)
	Ping()
	Dissect(b *bufio.Reader, isClient bool, tcpID *TcpID) interface{}
}
