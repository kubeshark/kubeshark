package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"plugin"
	"sync"
	"time"

	"github.com/google/martian/har"
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
	Dissect(b *bufio.Reader, isClient bool, tcpID *TcpID, counterPair *CounterPair, superTimer *SuperTimer, superIdentifier *SuperIdentifier, emitter Emitter, options *TrafficFilteringOptions) error
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
	ID                      uint `gorm:"primarykey"`
	CreatedAt               time.Time
	UpdatedAt               time.Time
	ProtocolName            string         `json:"protocolName" gorm:"column:protocolName"`
	ProtocolLongName        string         `json:"protocolLongName" gorm:"column:protocolLongName"`
	ProtocolAbbreviation    string         `json:"protocolAbbreviation" gorm:"column:protocolAbbreviation"`
	ProtocolVersion         string         `json:"protocolVersion" gorm:"column:protocolVersion"`
	ProtocolBackgroundColor string         `json:"protocolBackgroundColor" gorm:"column:protocolBackgroundColor"`
	ProtocolForegroundColor string         `json:"protocolForegroundColor" gorm:"column:protocolForegroundColor"`
	ProtocolFontSize        int8           `json:"protocolFontSize" gorm:"column:protocolFontSize"`
	ProtocolReferenceLink   string         `json:"protocolReferenceLink" gorm:"column:protocolReferenceLink"`
	Entry                   string         `json:"entry,omitempty" gorm:"column:entry"`
	EntryId                 string         `json:"entryId" gorm:"column:entryId"`
	Url                     string         `json:"url" gorm:"column:url"`
	Method                  string         `json:"method" gorm:"column:method"`
	Status                  int            `json:"status" gorm:"column:status"`
	RequestSenderIp         string         `json:"requestSenderIp" gorm:"column:requestSenderIp"`
	Service                 string         `json:"service" gorm:"column:service"`
	Timestamp               int64          `json:"timestamp" gorm:"column:timestamp"`
	ElapsedTime             int64          `json:"elapsedTime" gorm:"column:elapsedTime"`
	Path                    string         `json:"path" gorm:"column:path"`
	ResolvedSource          string         `json:"resolvedSource,omitempty" gorm:"column:resolvedSource"`
	ResolvedDestination     string         `json:"resolvedDestination,omitempty" gorm:"column:resolvedDestination"`
	SourceIp                string         `json:"sourceIp,omitempty" gorm:"column:sourceIp"`
	DestinationIp           string         `json:"destinationIp,omitempty" gorm:"column:destinationIp"`
	SourcePort              string         `json:"sourcePort,omitempty" gorm:"column:sourcePort"`
	DestinationPort         string         `json:"destinationPort,omitempty" gorm:"column:destinationPort"`
	IsOutgoing              bool           `json:"isOutgoing,omitempty" gorm:"column:isOutgoing"`
	ContractStatus          ContractStatus `json:"contractStatus,omitempty" gorm:"column:contractStatus"`
	ContractRequestReason   string         `json:"contractRequestReason,omitempty" gorm:"column:contractRequestReason"`
	ContractResponseReason  string         `json:"contractResponseReason,omitempty" gorm:"column:contractResponseReason"`
	ContractContent         string         `json:"contractContent,omitempty" gorm:"column:contractContent"`
	EstimatedSizeBytes      int            `json:"-" gorm:"column:estimatedSizeBytes"`
}

type MizuEntryWrapper struct {
	Protocol       Protocol                 `json:"protocol"`
	Representation string                   `json:"representation"`
	BodySize       int64                    `json:"bodySize"`
	Data           MizuEntry                `json:"data"`
	Rules          []map[string]interface{} `json:"rulesMatched,omitempty"`
	IsRulesEnabled bool                     `json:"isRulesEnabled"`
}

type BaseEntryDetails struct {
	Id              string          `json:"id,omitempty"`
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
	Latency         int64           `json:"latency"`
	Rules           ApplicableRules `json:"rules,omitempty"`
	ContractStatus  ContractStatus  `json:"contractStatus"`
}

type ApplicableRules struct {
	Latency             int64 `json:"latency,omitempty"`
	Status              bool  `json:"status,omitempty"`
	NumberOfRules       int   `json:"numberOfRules,omitempty"`
	NumberOfFailedRules int   `json:"numberOfFailedRules,omitempty"`
}

type ContractStatus int

type Contract struct {
	Status         ContractStatus `json:"status"`
	RequestReason  string         `json:"requestReason"`
	ResponseReason string         `json:"responseReason"`
	Content        string         `json:"content"`
}

type DataUnmarshaler interface {
	UnmarshalData(*MizuEntry) error
}

func (bed *BaseEntryDetails) UnmarshalData(entry *MizuEntry) error {
	bed.Protocol = Protocol{
		Name:            entry.ProtocolName,
		LongName:        entry.ProtocolLongName,
		Abbreviation:    entry.ProtocolAbbreviation,
		Version:         entry.ProtocolVersion,
		BackgroundColor: entry.ProtocolBackgroundColor,
		ForegroundColor: entry.ProtocolForegroundColor,
		FontSize:        entry.ProtocolFontSize,
		ReferenceLink:   entry.ProtocolReferenceLink,
	}
	bed.Id = entry.EntryId
	bed.Url = entry.Url
	bed.RequestSenderIp = entry.RequestSenderIp
	bed.Service = entry.Service
	bed.Path = entry.Path
	bed.Summary = entry.Path
	bed.StatusCode = entry.Status
	bed.Method = entry.Method
	bed.Timestamp = entry.Timestamp
	bed.SourceIp = entry.SourceIp
	bed.DestinationIp = entry.DestinationIp
	bed.SourcePort = entry.SourcePort
	bed.DestinationPort = entry.DestinationPort
	bed.IsOutgoing = entry.IsOutgoing
	bed.Latency = entry.ElapsedTime
	bed.ContractStatus = entry.ContractStatus
	return nil
}

const (
	TABLE string = "table"
	BODY  string = "body"
)

const (
	TypeHttpRequest = iota
	TypeHttpResponse
)

type HTTPPayload struct {
	Type uint8
	Data interface{}
}

type HTTPPayloader interface {
	MarshalJSON() ([]byte, error)
}

type HTTPWrapper struct {
	Method      string               `json:"method"`
	Url         string               `json:"url"`
	Details     interface{}          `json:"details"`
	RawRequest  *HTTPRequestWrapper  `json:"rawRequest"`
	RawResponse *HTTPResponseWrapper `json:"rawResponse"`
}

func (h HTTPPayload) MarshalJSON() ([]byte, error) {
	switch h.Type {
	case TypeHttpRequest:
		harRequest, err := har.NewRequest(h.Data.(*http.Request), true)
		if err != nil {
			return nil, errors.New("Failed converting request to HAR")
		}
		return json.Marshal(&HTTPWrapper{
			Method:     harRequest.Method,
			Url:        "",
			Details:    harRequest,
			RawRequest: &HTTPRequestWrapper{Request: h.Data.(*http.Request)},
		})
	case TypeHttpResponse:
		harResponse, err := har.NewResponse(h.Data.(*http.Response), true)
		if err != nil {
			return nil, errors.New("Failed converting response to HAR")
		}
		return json.Marshal(&HTTPWrapper{
			Method:      "",
			Url:         "",
			Details:     harResponse,
			RawResponse: &HTTPResponseWrapper{Response: h.Data.(*http.Response)},
		})
	default:
		panic(fmt.Sprintf("HTTP payload cannot be marshaled: %s\n", h.Type))
	}
}

type HTTPWrapperTricky struct {
	Method      string         `json:"method"`
	Url         string         `json:"url"`
	Details     interface{}    `json:"details"`
	RawRequest  *http.Request  `json:"rawRequest"`
	RawResponse *http.Response `json:"rawResponse"`
}

type HTTPMessage struct {
	IsRequest   bool              `json:"isRequest"`
	CaptureTime time.Time         `json:"captureTime"`
	Payload     HTTPWrapperTricky `json:"payload"`
}

type HTTPRequestResponsePair struct {
	Request  HTTPMessage `json:"request"`
	Response HTTPMessage `json:"response"`
}

type HTTPRequestWrapper struct {
	*http.Request
}

func (r *HTTPRequestWrapper) MarshalJSON() ([]byte, error) {
	body, _ := ioutil.ReadAll(r.Request.Body)
	r.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return json.Marshal(&struct {
		Body    string `json:"Body,omitempty"`
		GetBody string `json:"GetBody,omitempty"`
		Cancel  string `json:"Cancel,omitempty"`
		*http.Request
	}{
		Body:    string(body),
		Request: r.Request,
	})
}

type HTTPResponseWrapper struct {
	*http.Response
}

func (r *HTTPResponseWrapper) MarshalJSON() ([]byte, error) {
	body, _ := ioutil.ReadAll(r.Response.Body)
	r.Response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return json.Marshal(&struct {
		Body    string `json:"Body,omitempty"`
		GetBody string `json:"GetBody,omitempty"`
		Cancel  string `json:"Cancel,omitempty"`
		*http.Response
	}{
		Body:     string(body),
		Response: r.Response,
	})
}
