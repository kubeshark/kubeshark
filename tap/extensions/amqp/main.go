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

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, emitter api.Emitter) {
	// TODO: Implement
}

func (d dissecting) Analyze(item *api.OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *api.MizuEntry {
	// TODO: Implement
	return nil
}

func (d dissecting) Summarize(entry *api.MizuEntry) *api.BaseEntryDetails {
	// TODO: Implement
	return nil
}

var Dissector dissecting
