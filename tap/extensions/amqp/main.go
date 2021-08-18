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

func (d dissecting) Register(extension *api.Extension) {
	extension.Name = "amqp"
	extension.OutboundPorts = []string{"5671", "5672"}
	extension.InboundPorts = []string{}
}

func (d dissecting) Ping() {
	log.Printf("pong AMQP\n")
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, callback func(reqResPair *api.RequestResponsePair)) {
	// TODO: Implement
}

var Dissector dissecting
