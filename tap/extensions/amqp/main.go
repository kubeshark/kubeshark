package main

import (
	"bufio"
	"log"

	"github.com/up9inc/mizu/tap/api"
)

var protocol api.Protocol = api.Protocol{
	Name:            "amqp",
	LongName:        "Advanced Message Queuing Protocol",
	Abbreviation:    "AMQP",
	BackgroundColor: "#ff6600",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://www.rabbitmq.com/amqp-0-9-1-reference.html",
	OutboundPorts:   []string{"5671", "5672"},
	InboundPorts:    []string{},
}

func init() {
	log.Println("Initializing AMQP extension...")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s\n", protocol.Name)
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

func (d dissecting) Represent(entry string) ([]byte, error) {
	// TODO: Implement
	return nil, nil
}

var Dissector dissecting
