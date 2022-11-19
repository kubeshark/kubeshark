package api

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/kubeshark/kubeshark/tap/dbgctl"
)

const UnknownNamespace = ""

var UnknownIp = net.IP{0, 0, 0, 0}
var UnknownPort uint16 = 0

type ProtocolSummary struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Abbreviation string `json:"abbr"`
}

func (protocol *ProtocolSummary) ToString() string {
	return fmt.Sprintf("%s?%s?%s", protocol.Name, protocol.Version, protocol.Abbreviation)
}

func GetProtocolSummary(inputString string) *ProtocolSummary {
	splitted := strings.SplitN(inputString, "?", 3)
	return &ProtocolSummary{
		Name:         splitted[0],
		Version:      splitted[1],
		Abbreviation: splitted[2],
	}
}

type Protocol struct {
	ProtocolSummary
	LongName        string   `json:"longName"`
	Macro           string   `json:"macro"`
	BackgroundColor string   `json:"backgroundColor"`
	ForegroundColor string   `json:"foregroundColor"`
	FontSize        int8     `json:"fontSize"`
	ReferenceLink   string   `json:"referenceLink"`
	Ports           []string `json:"ports"`
	Priority        uint8    `json:"priority"`
}

type TCP struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
	Name string `json:"name"`
}

type Extension struct {
	Protocol  *Protocol
	Path      string
	Dissector Dissector
}

type Capture string

const (
	UndefinedCapture Capture = ""
	Pcap             Capture = "pcap"
	Envoy            Capture = "envoy"
	Linkerd          Capture = "linkerd"
	Ebpf             Capture = "ebpf"
)

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
	sync.Mutex
}

type GenericMessage struct {
	IsRequest   bool        `json:"isRequest"`
	CaptureTime time.Time   `json:"captureTime"`
	CaptureSize int         `json:"captureSize"`
	Payload     interface{} `json:"payload"`
}

type RequestResponsePair struct {
	Request  GenericMessage `json:"request"`
	Response GenericMessage `json:"response"`
}

type OutputChannelItem struct {
	// `Protocol` is modified in later stages of data propagation. Therefore, it's not a pointer.
	Protocol       Protocol
	Capture        Capture
	Timestamp      int64
	ConnectionInfo *ConnectionInfo
	Pair           *RequestResponsePair
	Namespace      string
}

type ReadProgress struct {
	readBytes   int
	lastCurrent int
}

func (p *ReadProgress) Feed(n int) {
	p.readBytes += n
}

func (p *ReadProgress) Current() (n int) {
	p.lastCurrent = p.readBytes - p.lastCurrent
	return p.lastCurrent
}

func (p *ReadProgress) Reset() {
	p.readBytes = 0
	p.lastCurrent = 0
}

type Dissector interface {
	Register(*Extension)
	GetProtocols() map[string]*Protocol
	Ping()
	Dissect(b *bufio.Reader, reader TcpReader, options *TrafficFilteringOptions) error
	Analyze(item *OutputChannelItem, resolvedSource string, resolvedDestination string, namespace string) *Entry
	Summarize(entry *Entry) *BaseEntry
	Represent(request map[string]interface{}, response map[string]interface{}) (object []byte, err error)
	Macros() map[string]string
	NewResponseRequestMatcher() RequestResponseMatcher
}

type RequestResponseMatcher interface {
	GetMap() *sync.Map
	SetMaxTry(value int)
}

type Emitting struct {
	AppStats      *AppStats
	OutputChannel chan *OutputChannelItem
}

type Emitter interface {
	Emit(item *OutputChannelItem)
}

func (e *Emitting) Emit(item *OutputChannelItem) {
	e.AppStats.IncMatchedPairs()

	if dbgctl.KubesharkTapperDisableEmitting {
		return
	}

	e.OutputChannel <- item
}

type Entry struct {
	Id           string                 `json:"id"`
	Protocol     ProtocolSummary        `json:"protocol"`
	Capture      Capture                `json:"capture"`
	Source       *TCP                   `json:"src"`
	Destination  *TCP                   `json:"dst"`
	Namespace    string                 `json:"namespace"`
	Outgoing     bool                   `json:"outgoing"`
	Timestamp    int64                  `json:"timestamp"`
	StartTime    time.Time              `json:"startTime"`
	Request      map[string]interface{} `json:"request"`
	Response     map[string]interface{} `json:"response"`
	RequestSize  int                    `json:"requestSize"`
	ResponseSize int                    `json:"responseSize"`
	ElapsedTime  int64                  `json:"elapsedTime"`
}

type EntryWrapper struct {
	Protocol       Protocol   `json:"protocol"`
	Representation string     `json:"representation"`
	Data           *Entry     `json:"data"`
	Base           *BaseEntry `json:"base"`
}

type BaseEntry struct {
	Id           string   `json:"id"`
	Protocol     Protocol `json:"proto,omitempty"`
	Capture      Capture  `json:"capture"`
	Summary      string   `json:"summary,omitempty"`
	SummaryQuery string   `json:"summaryQuery,omitempty"`
	Status       int      `json:"status"`
	StatusQuery  string   `json:"statusQuery"`
	Method       string   `json:"method,omitempty"`
	MethodQuery  string   `json:"methodQuery,omitempty"`
	Timestamp    int64    `json:"timestamp,omitempty"`
	Source       *TCP     `json:"src"`
	Destination  *TCP     `json:"dst"`
	IsOutgoing   bool     `json:"isOutgoing,omitempty"`
	Latency      int64    `json:"latency"`
}

const (
	TABLE string = "table"
	BODY  string = "body"
)

type SectionData struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Data     string `json:"data"`
	Encoding string `json:"encoding,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Selector string `json:"selector,omitempty"`
}

type TableData struct {
	Name     string      `json:"name"`
	Value    interface{} `json:"value"`
	Selector string      `json:"selector"`
}

type TcpReaderDataMsg interface {
	GetBytes() []byte
	GetTimestamp() time.Time
}

type TcpReader interface {
	Read(p []byte) (int, error)
	GetReqResMatcher() RequestResponseMatcher
	GetIsClient() bool
	GetReadProgress() *ReadProgress
	GetParent() TcpStream
	GetTcpID() *TcpID
	GetCounterPair() *CounterPair
	GetCaptureTime() time.Time
	GetEmitter() Emitter
	GetIsClosed() bool
}

type TcpStream interface {
	SetProtocol(protocol *Protocol)
	GetOrigin() Capture
	GetReqResMatchers() []RequestResponseMatcher
	GetIsTapTarget() bool
	GetIsClosed() bool
}

type TcpStreamMap interface {
	Range(f func(key, value interface{}) bool)
	Store(key, value interface{})
	Delete(key interface{})
	NextId() int64
	CloseTimedoutTcpStreamChannels()
}
