package main

import (
	"bufio"
	"log"

	"github.com/up9inc/mizu/tap/api"
)

func init() {
	log.Println("Initializing AMQP extension.")
}

type dissecting string

func (g dissecting) Register(extension *api.Extension) {
	extension.Name = "amqp"
	extension.OutboundPorts = []string{"5671", "5672"}
	extension.InboundPorts = []string{}
}

func (g dissecting) Ping() {
	log.Printf("pong AMQP\n")
}

func (g dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID) *api.OutputChannelItem {
	// TODO: Implement
	return nil
}

var Dissector dissecting
