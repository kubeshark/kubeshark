package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/google/martian/har"
)

const mizuTestEnvVar = "MIZU_TEST"

type Protocol struct {
	Name            string   `json:"name"`
	LongName        string   `json:"longName"`
	Abbreviation    string   `json:"abbr"`
	Macro           string   `json:"macro"`
	Version         string   `json:"version"`
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
	Summary        *BaseEntry
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
	Dissect(b *bufio.Reader, isClient bool, tcpID *TcpID, counterPair *CounterPair, superTimer *SuperTimer, superIdentifier *SuperIdentifier, emitter Emitter, options *TrafficFilteringOptions, reqResMatcher interface{}) error
	Analyze(item *OutputChannelItem, resolvedSource string, resolvedDestination string) *Entry
	Represent(request map[string]interface{}, response map[string]interface{}) (object []byte, bodySize int64, err error)
	Macros() map[string]string
	NewResponseRequestMatcher() interface{}
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

type Entry struct {
	Id                     uint                   `json:"id"`
	Protocol               Protocol               `json:"proto"`
	Source                 *TCP                   `json:"src"`
	Destination            *TCP                   `json:"dst"`
	Outgoing               bool                   `json:"outgoing"`
	Timestamp              int64                  `json:"timestamp"`
	StartTime              time.Time              `json:"startTime"`
	Request                map[string]interface{} `json:"request"`
	Response               map[string]interface{} `json:"response"`
	Summary                string                 `json:"summary"`
	Method                 string                 `json:"method"`
	Status                 int                    `json:"status"`
	ElapsedTime            int64                  `json:"elapsedTime"`
	Path                   string                 `json:"path"`
	IsOutgoing             bool                   `json:"isOutgoing,omitempty"`
	Rules                  ApplicableRules        `json:"rules,omitempty"`
	ContractStatus         ContractStatus         `json:"contractStatus,omitempty"`
	ContractRequestReason  string                 `json:"contractRequestReason,omitempty"`
	ContractResponseReason string                 `json:"contractResponseReason,omitempty"`
	ContractContent        string                 `json:"contractContent,omitempty"`
	HTTPPair               string                 `json:"httpPair,omitempty"`
}

type EntryWrapper struct {
	Protocol       Protocol                 `json:"protocol"`
	Representation string                   `json:"representation"`
	BodySize       int64                    `json:"bodySize"`
	Data           *Entry                   `json:"data"`
	Rules          []map[string]interface{} `json:"rulesMatched,omitempty"`
	IsRulesEnabled bool                     `json:"isRulesEnabled"`
}

type BaseEntry struct {
	Id             uint            `json:"id"`
	Protocol       Protocol        `json:"proto,omitempty"`
	Url            string          `json:"url,omitempty"`
	Path           string          `json:"path,omitempty"`
	Summary        string          `json:"summary,omitempty"`
	StatusCode     int             `json:"status"`
	Method         string          `json:"method,omitempty"`
	Timestamp      int64           `json:"timestamp,omitempty"`
	Source         *TCP            `json:"src"`
	Destination    *TCP            `json:"dst"`
	IsOutgoing     bool            `json:"isOutgoing,omitempty"`
	Latency        int64           `json:"latency"`
	Rules          ApplicableRules `json:"rules,omitempty"`
	ContractStatus ContractStatus  `json:"contractStatus"`
}

type ApplicableRules struct {
	Latency       int64 `json:"latency,omitempty"`
	Status        bool  `json:"status,omitempty"`
	NumberOfRules int   `json:"numberOfRules,omitempty"`
}

type ContractStatus int

type Contract struct {
	Status         ContractStatus `json:"status"`
	RequestReason  string         `json:"requestReason"`
	ResponseReason string         `json:"responseReason"`
	Content        string         `json:"content"`
}

func Summarize(entry *Entry) *BaseEntry {
	return &BaseEntry{
		Id:             entry.Id,
		Protocol:       entry.Protocol,
		Path:           entry.Path,
		Summary:        entry.Summary,
		StatusCode:     entry.Status,
		Method:         entry.Method,
		Timestamp:      entry.Timestamp,
		Source:         entry.Source,
		Destination:    entry.Destination,
		IsOutgoing:     entry.IsOutgoing,
		Latency:        entry.ElapsedTime,
		Rules:          entry.Rules,
		ContractStatus: entry.ContractStatus,
	}
}

type DataUnmarshaler interface {
	UnmarshalData(*Entry) error
}

func (bed *BaseEntry) UnmarshalData(entry *Entry) error {
	bed.Protocol = entry.Protocol
	bed.Id = entry.Id
	bed.Path = entry.Path
	bed.Summary = entry.Summary
	bed.StatusCode = entry.Status
	bed.Method = entry.Method
	bed.Timestamp = entry.Timestamp
	bed.Source = entry.Source
	bed.Destination = entry.Destination
	bed.IsOutgoing = entry.IsOutgoing
	bed.Latency = entry.ElapsedTime
	bed.ContractStatus = entry.ContractStatus
	return nil
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
	_, testEnvEnabled := os.LookupEnv(mizuTestEnvVar)
	switch h.Type {
	case TypeHttpRequest:
		harRequest, err := har.NewRequest(h.Data.(*http.Request), true)
		if err != nil {
			return nil, errors.New("Failed converting request to HAR")
		}
		sort.Slice(harRequest.Headers, func(i, j int) bool {
			if harRequest.Headers[i].Name < harRequest.Headers[j].Name {
				return true
			}
			if harRequest.Headers[i].Name > harRequest.Headers[j].Name {
				return false
			}
			return harRequest.Headers[i].Value < harRequest.Headers[j].Value
		})
		sort.Slice(harRequest.QueryString, func(i, j int) bool {
			if harRequest.QueryString[i].Name < harRequest.QueryString[j].Name {
				return true
			}
			if harRequest.QueryString[i].Name > harRequest.QueryString[j].Name {
				return false
			}
			return harRequest.QueryString[i].Value < harRequest.QueryString[j].Value
		})
		if harRequest.PostData != nil {
			sort.Slice(harRequest.PostData.Params, func(i, j int) bool {
				if harRequest.PostData.Params[i].Name < harRequest.PostData.Params[j].Name {
					return true
				}
				if harRequest.PostData.Params[i].Name > harRequest.PostData.Params[j].Name {
					return false
				}
				return harRequest.PostData.Params[i].Value < harRequest.PostData.Params[j].Value
			})
		}
		if testEnvEnabled {
			harRequest.URL = ""
		}
		var reqWrapper *HTTPRequestWrapper
		if !testEnvEnabled {
			reqWrapper = &HTTPRequestWrapper{Request: h.Data.(*http.Request)}
		}
		return json.Marshal(&HTTPWrapper{
			Method:     harRequest.Method,
			Details:    harRequest,
			RawRequest: reqWrapper,
		})
	case TypeHttpResponse:
		harResponse, err := har.NewResponse(h.Data.(*http.Response), true)
		if err != nil {
			return nil, errors.New("Failed converting response to HAR")
		}
		sort.Slice(harResponse.Headers, func(i, j int) bool {
			if harResponse.Headers[i].Name < harResponse.Headers[j].Name {
				return true
			}
			if harResponse.Headers[i].Name > harResponse.Headers[j].Name {
				return false
			}
			return harResponse.Headers[i].Value < harResponse.Headers[j].Value
		})
		sort.Slice(harResponse.Cookies, func(i, j int) bool {
			if harResponse.Cookies[i].Name < harResponse.Cookies[j].Name {
				return true
			}
			if harResponse.Cookies[i].Name > harResponse.Cookies[j].Name {
				return false
			}
			return harResponse.Cookies[i].Value < harResponse.Cookies[j].Value
		})
		var resWrapper *HTTPResponseWrapper
		if !testEnvEnabled {
			resWrapper = &HTTPResponseWrapper{Response: h.Data.(*http.Response)}
		}
		return json.Marshal(&HTTPWrapper{
			Method:      "",
			Url:         "",
			Details:     harResponse,
			RawResponse: resWrapper,
		})
	default:
		panic(fmt.Sprintf("HTTP payload cannot be marshaled: %v", h.Type))
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
	return json.Marshal(&struct { //nolint
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
	return json.Marshal(&struct { //nolint
		Body    string `json:"Body,omitempty"`
		GetBody string `json:"GetBody,omitempty"`
		Cancel  string `json:"Cancel,omitempty"`
		*http.Response
	}{
		Body:     string(body),
		Response: r.Response,
	})
}
