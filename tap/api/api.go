package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/google/martian/har"

	"github.com/up9inc/mizu/tap/dbgctl"
)

const mizuTestEnvVar = "MIZU_TEST"
const UNKNOWN_NAMESPACE = ""

var UnknownIp net.IP = net.IP{0, 0, 0, 0}
var UnknownPort uint16 = 0

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
	Summary        *BaseEntry
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

	if dbgctl.MizuTapperDisableEmitting {
		return
	}

	e.OutputChannel <- item
}

type Entry struct {
	Id                     string                 `json:"id"`
	Protocol               Protocol               `json:"proto"`
	Capture                Capture                `json:"capture"`
	Source                 *TCP                   `json:"src"`
	Destination            *TCP                   `json:"dst"`
	Namespace              string                 `json:"namespace"`
	Outgoing               bool                   `json:"outgoing"`
	Timestamp              int64                  `json:"timestamp"`
	StartTime              time.Time              `json:"startTime"`
	Request                map[string]interface{} `json:"request"`
	Response               map[string]interface{} `json:"response"`
	RequestSize            int                    `json:"requestSize"`
	ResponseSize           int                    `json:"responseSize"`
	ElapsedTime            int64                  `json:"elapsedTime"`
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
	Data           *Entry                   `json:"data"`
	Base           *BaseEntry               `json:"base"`
	Rules          []map[string]interface{} `json:"rulesMatched,omitempty"`
	IsRulesEnabled bool                     `json:"isRulesEnabled"`
}

type BaseEntry struct {
	Id             string          `json:"id"`
	Protocol       Protocol        `json:"proto,omitempty"`
	Capture        Capture         `json:"capture"`
	Summary        string          `json:"summary,omitempty"`
	SummaryQuery   string          `json:"summaryQuery,omitempty"`
	Status         int             `json:"status"`
	StatusQuery    string          `json:"statusQuery"`
	Method         string          `json:"method,omitempty"`
	MethodQuery    string          `json:"methodQuery,omitempty"`
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
