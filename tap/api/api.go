package api

import (
	"bufio"
	"log"
	"plugin"
	"time"
)

type Protocol struct {
	Name            string   `json:"name"`
	LongName        string   `json:"long_name"`
	Abbreviation    string   `json:"abbreviation"`
	BackgroundColor string   `json:"background_color"`
	ForegroundColor string   `json:"foreground_color"`
	FontSize        int8     `json:"font_size"`
	ReferenceLink   string   `json:"reference_link"`
	OutboundPorts   []string `json:"outbound_ports"`
	InboundPorts    []string `json:"inbound_ports"`
}

type Extension struct {
	Protocol  Protocol
	Path      string
	Plug      *plugin.Plugin
	Dissector Dissector
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
	IsRequest   bool        `json:"is_request"`
	CaptureTime time.Time   `json:"capture_time"`
	Payload     interface{} `json:"payload"`
}

type RequestResponsePair struct {
	Request  GenericMessage `json:"request"`
	Response GenericMessage `json:"response"`
}

type OutputChannelItem struct {
	Protocol       Protocol
	Timestamp      int64
	ConnectionInfo *ConnectionInfo
	Pair           *RequestResponsePair
}

type Dissector interface {
	Register(*Extension)
	Ping()
	Dissect(b *bufio.Reader, isClient bool, tcpID *TcpID, emitter Emitter)
	Analyze(item *OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *MizuEntry
	Summarize(entry *MizuEntry) *BaseEntryDetails
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

type MizuEntry struct {
	ID                  uint `gorm:"primarykey"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Entry               string `json:"entry,omitempty" gorm:"column:entry"`
	EntryId             string `json:"entryId" gorm:"column:entryId"`
	Url                 string `json:"url" gorm:"column:url"`
	Method              string `json:"method" gorm:"column:method"`
	Status              int    `json:"status" gorm:"column:status"`
	RequestSenderIp     string `json:"requestSenderIp" gorm:"column:requestSenderIp"`
	Service             string `json:"service" gorm:"column:service"`
	Timestamp           int64  `json:"timestamp" gorm:"column:timestamp"`
	Path                string `json:"path" gorm:"column:path"`
	ResolvedSource      string `json:"resolvedSource,omitempty" gorm:"column:resolvedSource"`
	ResolvedDestination string `json:"resolvedDestination,omitempty" gorm:"column:resolvedDestination"`
	SourceIp            string `json:"sourceIp,omitempty" gorm:"column:sourceIp"`
	DestinationIp       string `json:"destinationIp,omitempty" gorm:"column:destinationIp"`
	SourcePort          string `json:"sourcePort,omitempty" gorm:"column:sourcePort"`
	DestinationPort     string `json:"destinationPort,omitempty" gorm:"column:destinationPort"`
	IsOutgoing          bool   `json:"isOutgoing,omitempty" gorm:"column:isOutgoing"`
	EstimatedSizeBytes  int    `json:"-" gorm:"column:estimatedSizeBytes"`
}

type BaseEntryDetails struct {
	Id              string          `json:"id,omitempty"`
	Protocol        Protocol        `json:"protocol,omitempty"`
	Url             string          `json:"url,omitempty"`
	RequestSenderIp string          `json:"request_sender_ip,omitempty"`
	Service         string          `json:"service,omitempty"`
	Summary         string          `json:"summary,omitempty"`
	StatusCode      int             `json:"status_code,omitempty"`
	Method          string          `json:"method,omitempty"`
	Timestamp       int64           `json:"timestamp,omitempty"`
	SourceIp        string          `json:"source_ip,omitempty"`
	DestinationIp   string          `json:"destination_ip,omitempty"`
	SourcePort      string          `json:"source_port,omitempty"`
	DestinationPort string          `json:"destination_port,omitempty"`
	IsOutgoing      bool            `json:"isOutgoing,omitempty"`
	Latency         int64           `json:"latency,omitempty"`
	Rules           ApplicableRules `json:"rules,omitempty"`
}

type ApplicableRules struct {
	Latency int64 `json:"latency,omitempty"`
	Status  bool  `json:"status,omitempty"`
}

type DataUnmarshaler interface {
	UnmarshalData(*MizuEntry) error
}

func (bed *BaseEntryDetails) UnmarshalData(entry *MizuEntry) error {
	entryUrl := entry.Url
	service := entry.Service
	bed.Id = entry.EntryId
	bed.Url = entryUrl
	bed.Service = service
	bed.Summary = entry.Path
	bed.StatusCode = entry.Status
	bed.Method = entry.Method
	bed.Timestamp = entry.Timestamp
	bed.RequestSenderIp = entry.RequestSenderIp
	bed.IsOutgoing = entry.IsOutgoing
	return nil
}
