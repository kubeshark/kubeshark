package main

import (
	"bufio"
	"log"

	"github.com/up9inc/mizu/tap/api"
)

var _protocol api.Protocol = api.Protocol{
	Name:            "kafka",
	LongName:        "Apache Kafka Protocol",
	Abbreviation:    "KAFKA",
	BackgroundColor: "#000000",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://kafka.apache.org/protocol",
	Ports:           []string{"9092"},
}

func init() {
	log.Println("Initializing Kafka extension...")
}

type dissecting string

func (d dissecting) Register(extension *api.Extension) {
	extension.Protocol = _protocol
}

func (d dissecting) Ping() {
	log.Printf("pong %s\n", _protocol.Name)
}

func (d dissecting) Dissect(b *bufio.Reader, isClient bool, tcpID *api.TcpID, emitter api.Emitter) {
	if isClient {
		ReadRequest(b, tcpID)
	} else {
		ReadResponse(b, tcpID, emitter)
	}
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
