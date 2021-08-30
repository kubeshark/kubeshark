package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	tapApi "github.com/up9inc/mizu/tap/api"

	"mizuserver/pkg/rules"
	"mizuserver/pkg/utils"
	"time"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
)

func GetEntry(r *tapApi.MizuEntry, v tapApi.DataUnmarshaler) error {
	return v.UnmarshalData(r)
}

func NewApplicableRules(status bool, latency int64, number int) tapApi.ApplicableRules {
	ar := tapApi.ApplicableRules{}
	ar.Status = status
	ar.Latency = latency
	ar.NumberOfRules = number
	return ar
}

type FullEntryDetails struct {
	har.Entry
}

type FullEntryDetailsExtra struct {
	har.Entry
}

func (fed *FullEntryDetails) UnmarshalData(entry *tapApi.MizuEntry) error {
	if err := json.Unmarshal([]byte(entry.Entry), &fed.Entry); err != nil {
		return err
	}

	if entry.ResolvedDestination != "" {
		fed.Entry.Request.URL = utils.SetHostname(fed.Entry.Request.URL, entry.ResolvedDestination)
	}
	return nil
}

func (fedex *FullEntryDetailsExtra) UnmarshalData(entry *tapApi.MizuEntry) error {
	if err := json.Unmarshal([]byte(entry.Entry), &fedex.Entry); err != nil {
		return err
	}

	if entry.ResolvedSource != "" {
		fedex.Entry.Request.Headers = append(fedex.Request.Headers, har.Header{Name: "x-mizu-source", Value: entry.ResolvedSource})
	}
	if entry.ResolvedDestination != "" {
		fedex.Entry.Request.Headers = append(fedex.Request.Headers, har.Header{Name: "x-mizu-destination", Value: entry.ResolvedDestination})
		fedex.Entry.Request.URL = utils.SetHostname(fedex.Entry.Request.URL, entry.ResolvedDestination)
	}
	return nil
}

type EntriesFilter struct {
	Limit     int    `form:"limit" validate:"required,min=1,max=200"`
	Operator  string `form:"operator" validate:"required,oneof='lt' 'gt'"`
	Timestamp int64  `form:"timestamp" validate:"required,min=1"`
}

type UploadEntriesRequestQuery struct {
	Dest             string `form:"dest"`
	SleepIntervalSec int    `form:"interval"`
}

type HarFetchRequestQuery struct {
	From int64 `form:"from"`
	To   int64 `form:"to"`
}

type WebSocketEntryMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tapApi.BaseEntryDetails `json:"data,omitempty"`
}

type WebSocketTappedEntryMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tapApi.OutputChannelItem
}

type WebsocketOutboundLinkMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tap.OutboundLink
}

func CreateBaseEntryWebSocketMessage(base *tapApi.BaseEntryDetails) ([]byte, error) {
	message := &WebSocketEntryMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebSocketMessageTypeEntry,
		},
		Data: base,
	}
	return json.Marshal(message)
}

func CreateWebsocketTappedEntryMessage(base *tapApi.OutputChannelItem) ([]byte, error) {
	message := &WebSocketTappedEntryMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebSocketMessageTypeTappedEntry,
		},
		Data: base,
	}
	return json.Marshal(message)
}

func CreateWebsocketOutboundLinkMessage(base *tap.OutboundLink) ([]byte, error) {
	message := &WebsocketOutboundLinkMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebsocketMessageTypeOutboundLink,
		},
		Data: base,
	}
	return json.Marshal(message)
}

// ExtendedHAR is the top level object of a HAR log.
type ExtendedHAR struct {
	Log *ExtendedLog `json:"log"`
}

// ExtendedLog is the HAR HTTP request and response log.
type ExtendedLog struct {
	// Version number of the HAR format.
	Version string `json:"version"`
	// Creator holds information about the log creator application.
	Creator *ExtendedCreator `json:"creator"`
	// Entries is a list containing requests and responses.
	Entries []*har.Entry `json:"entries"`
}

type ExtendedCreator struct {
	*har.Creator
	Source *string `json:"_source"`
}

type FullEntryWithPolicy struct {
	RulesMatched []rules.RulesMatched `json:"rulesMatched,omitempty"`
	Entry        har.Entry            `json:"entry"`
	Service      string               `json:"service"`
}

func (fewp *FullEntryWithPolicy) UnmarshalData(entry *tapApi.MizuEntry) error {
	if err := json.Unmarshal([]byte(entry.Entry), &fewp.Entry); err != nil {
		return err
	}

	_, resultPolicyToSend := rules.MatchRequestPolicy(fewp.Entry, entry.Service)
	fewp.RulesMatched = resultPolicyToSend
	fewp.Service = entry.Service
	return nil
}

func RunValidationRulesState(harEntry har.Entry, service string) tapApi.ApplicableRules {
	numberOfRules, resultPolicyToSend := rules.MatchRequestPolicy(harEntry, service)
	statusPolicyToSend, latency, numberOfRules := rules.PassedValidationRules(resultPolicyToSend, numberOfRules)
	ar := NewApplicableRules(statusPolicyToSend, latency, numberOfRules)
	return ar
}

func NewEntry(request *http.Request, requestTime time.Time, response *http.Response, responseTime time.Time) (*har.Entry, error) {
	harRequest, err := har.NewRequest(request, false)
	if err != nil {
		fmt.Printf("Failed converting request to HAR %s (%v,%+v)", err, err, err)
		return nil, errors.New("failed converting request to HAR")
	}

	// For requests with multipart/form-data or application/x-www-form-urlencoded Content-Type,
	// martian/har will parse the request body and place the parameters in harRequest.PostData.Params
	// instead of harRequest.PostData.Text (as the HAR spec requires it).
	// Mizu currently only looks at PostData.Text. Therefore, instead of letting martian/har set the content of
	// PostData, always copy the request body to PostData.Text.
	if request.ContentLength > 0 {
		reqBody, err := ioutil.ReadAll(request.Body)
		if err != nil {
			fmt.Printf("Failed converting request to HAR %s (%v,%+v)", err, err, err)
			return nil, errors.New("failed reading request body")
		}
		request.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
		harRequest.PostData.Text = string(reqBody)
	}

	harResponse, err := har.NewResponse(response, true)
	if err != nil {
		fmt.Printf("Failed converting response to HAR %s (%v,%+v)", err, err, err)
		return nil, errors.New("failed converting response to HAR")
	}

	if harRequest.PostData != nil && strings.HasPrefix(harRequest.PostData.MimeType, "application/grpc") {
		// Force HTTP/2 gRPC into HAR template

		harRequest.URL = fmt.Sprintf("%s://%s%s", request.Header.Get(":scheme"), request.Header.Get(":authority"), request.Header.Get(":path"))

		status, err := strconv.Atoi(response.Header.Get(":status"))
		if err != nil {
			fmt.Printf("Failed converting status to int %s (%v,%+v)", err, err, err)
			return nil, errors.New("failed converting response status to int for HAR")
		}
		harResponse.Status = status
	} else {
		// Martian copies http.Request.URL.String() to har.Request.URL, which usually contains the path.
		// However, according to the HAR spec, the URL field needs to be the absolute URL.
		var scheme string
		if request.URL.Scheme != "" {
			scheme = request.URL.Scheme
		} else {
			scheme = "http"
		}
		harRequest.URL = fmt.Sprintf("%s://%s%s", scheme, request.Host, request.URL)
	}

	totalTime := responseTime.Sub(requestTime).Round(time.Millisecond).Milliseconds()
	if totalTime < 1 {
		totalTime = 1
	}

	harEntry := har.Entry{
		StartedDateTime: time.Now().UTC(),
		Time:            totalTime,
		Request:         harRequest,
		Response:        harResponse,
		Cache:           &har.Cache{},
		Timings: &har.Timings{
			Send:    -1,
			Wait:    -1,
			Receive: totalTime,
		},
	}

	return &harEntry, nil
}
