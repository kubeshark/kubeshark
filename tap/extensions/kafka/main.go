package main

import (
	"bufio"
	"log"

	"github.com/up9inc/mizu/tap/api"
)

func init() {
	log.Println("Initializing Kafka extension.")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Name = "kafka"
	extension.OutboundPorts = []string{"9092"}
	extension.InboundPorts = []string{}
}

func (d dissecting) Ping() {
	log.Printf("pong Kafka\n")
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, callback func(item *api.OutputChannelItem)) {
	// TODO: Implement
}

var Dissector dissecting
