package api

import (
	"bufio"
	"plugin"
	"sync"
	"time"
)

type Protocol struct {
	Name            string   `json:"name"`
	LongName        string   `json:"longName"`
	Abbreviation    string   `json:"abbreviation"`
	Version         string   `json:"version"`
	BackgroundColor string   `json:"backgroundColor"`
	ForegroundColor string   `json:"foregroundColor"`
	FontSize        int8     `json:"fontSize"`
	ReferenceLink   string   `json:"referenceLink"`
	Ports           []string `json:"ports"`
	Priority        uint8    `json:"priority"`
}

type Extension struct {
	Protocol   *Protocol
	Path       string
	Plug       *plugin.Plugin
	Dissector  Dissector
	MatcherMap *sync.Map
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

type CounterPair struct {
	Request  uint
	Response uint
}

type GenericMessage struct {
	IsRequest   bool        `json:"isRequest"`
	CaptureTime time.Time   `json:"captureTime"`
	Payload     interface{} `json:"payload"`
}

type RequestResponsePair struct {
	Request  GenericMessage `json:"request"`
	Response GenericMessage `json:"response"`
}

// `Protocol` is modified in the later stages of data propagation. Therefore it's not a pointer.
type OutputChannelItem struct {
	Protocol       Protocol
	Timestamp      int64
	ConnectionInfo *ConnectionInfo
	Pair           *RequestResponsePair
	Summary        *BaseEntryDetails
}

type SuperTimer struct {
	CaptureTime time.Time
}

type SuperIdentifier struct {
	Protocol       *Protocol
	IsClosedOthers bool
}

type Dissector interface {
	Register(*Extension)
	Ping()
	Dissect(b *bufio.Reader, isClient bool, tcpID *TcpID, counterPair *CounterPair, superTimer *SuperTimer, superIdentifier *SuperIdentifier, emitter Emitter) error
	Analyze(item *OutputChannelItem, entryId string, resolvedSource string, resolvedDestination string) *MizuEntry
	Summarize(entry *MizuEntry) *BaseEntryDetails
	Represent(entry *MizuEntry) (protocol Protocol, object []byte, bodySize int64, err error)
}

type Emitting struct {
	AppStats      *AppStats
	OutputChannel chan *OutputChannelItem
}

type Emitter interface {
	Emit(item *OutputChannelItem)
}

func (e *Emitting) Emit(item *OutputChannelItem) {
	e.OutputChannel <- item
	e.AppStats.IncMatchedPairs()
}

type MizuEntry struct {
	Id                  uint              `json:"id"`
	Protocol            Protocol          `json:"proto"`
	Timestamp           int64             `json:"timestamp"`
	Request             interface{}       `json:"request"`
	Response            interface{}       `json:"response"`
	Summary             *BaseEntryDetails `json:"summary"`
	Entry               string            `json:"entry,omitempty"`
	EntryId             string            `json:"entryId" gorm:"column:entryId"`
	Url                 string            `json:"url" gorm:"column:url"`
	Method              string            `json:"method" gorm:"column:method"`
	Status              int               `json:"status" gorm:"column:status"`
	RequestSenderIp     string            `json:"requestSenderIp" gorm:"column:requestSenderIp"`
	Service             string            `json:"service" gorm:"column:service"`
	ElapsedTime         int64             `json:"elapsedTime" gorm:"column:elapsedTime"`
	Path                string            `json:"path" gorm:"column:path"`
	ResolvedSource      string            `json:"resolvedSource,omitempty" gorm:"column:resolvedSource"`
	ResolvedDestination string            `json:"resolvedDestination,omitempty" gorm:"column:resolvedDestination"`
	SourceIp            string            `json:"sourceIp,omitempty" gorm:"column:sourceIp"`
	DestinationIp       string            `json:"destinationIp,omitempty" gorm:"column:destinationIp"`
	SourcePort          string            `json:"sourcePort,omitempty" gorm:"column:sourcePort"`
	DestinationPort     string            `json:"destinationPort,omitempty" gorm:"column:destinationPort"`
	IsOutgoing          bool              `json:"isOutgoing,omitempty" gorm:"column:isOutgoing"`
	EstimatedSizeBytes  int               `json:"-" gorm:"column:estimatedSizeBytes"`
}

type MizuEntryWrapper struct {
	Protocol       Protocol  `json:"protocol"`
	Representation string    `json:"representation"`
	BodySize       int64     `json:"bodySize"`
	Data           MizuEntry `json:"data"`
}

type BaseEntryDetails struct {
	Id              uint            `json:"id"`
	Protocol        Protocol        `json:"protocol,omitempty"`
	Url             string          `json:"url,omitempty"`
	RequestSenderIp string          `json:"requestSenderIp,omitempty"`
	Service         string          `json:"service,omitempty"`
	Path            string          `json:"path,omitempty"`
	Summary         string          `json:"summary,omitempty"`
	StatusCode      int             `json:"statusCode"`
	Method          string          `json:"method,omitempty"`
	Timestamp       int64           `json:"timestamp,omitempty"`
	SourceIp        string          `json:"sourceIp,omitempty"`
	DestinationIp   string          `json:"destinationIp,omitempty"`
	SourcePort      string          `json:"sourcePort,omitempty"`
	DestinationPort string          `json:"destinationPort,omitempty"`
	IsOutgoing      bool            `json:"isOutgoing,omitempty"`
	Latency         int64           `json:"latency,omitempty"`
	Rules           ApplicableRules `json:"rules,omitempty"`
}

type ApplicableRules struct {
	Latency       int64 `json:"latency,omitempty"`
	Status        bool  `json:"status,omitempty"`
	NumberOfRules int   `json:"numberOfRules,omitempty"`
}

type DataUnmarshaler interface {
	UnmarshalData(*MizuEntry) error
}

func (bed *BaseEntryDetails) UnmarshalData(entry *MizuEntry) error {
	bed.Protocol = entry.Protocol
	bed.Id = entry.Id
	bed.Url = entry.Url
	bed.Service = entry.Service
	bed.Summary = entry.Path
	bed.StatusCode = entry.Status
	bed.Method = entry.Method
	bed.Timestamp = entry.Timestamp
	bed.RequestSenderIp = entry.RequestSenderIp
	bed.IsOutgoing = entry.IsOutgoing
	return nil
}

const (
	TABLE string = "table"
	BODY  string = "body"
)
