package models

import (
	"encoding/json"
	"time"

	tapApi "github.com/up9inc/mizu/tap/api"

	"mizuserver/pkg/rules"
	"mizuserver/pkg/utils"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
)

func GetEntry(r *tapApi.MizuEntry, v tapApi.DataUnmarshaler) error {
	return v.UnmarshalData(r)
}

//func NewApplicableRules(status bool, latency int64, number int) tapApi.ApplicableRules {
//	ar := tapApi.ApplicableRules{}
//	ar.Status = status
//	ar.Latency = latency
//	ar.NumberOfRules = number
//	return ar
//}

type FullEntryDetails struct {
	har.Entry
}

type FullEntryDetailsExtra struct {
	har.Entry
}

func NewEntry(startDate time.Time, harRequest *har.Request, harResponse *har.Response, totalTime int64) har.Entry {
	return har.Entry{
		StartedDateTime: startDate,
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
}


func (fed *FullEntryDetails) UnmarshalData(entry *tapApi.MizuEntry) error {
	var root tapApi.RequestResponsePair
	err := json.Unmarshal([]byte(entry.Entry), &root)
	if err != nil {
		return err
	}
	requestPayload := root.Request.Payload.(map[string]interface{})
	responsePayload := root.Response.Payload.(map[string]interface{})
	totalTime := root.Response.CaptureTime.Sub(root.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	fed.Entry = NewEntry(root.Request.CaptureTime, requestPayload["details"].(interface{}).(*har.Request), responsePayload["details"].(interface{}).(*har.Response), totalTime)


	if entry.ResolvedDestination != "" {
		fed.Entry.Request.URL = utils.SetHostname(fed.Entry.Request.URL, entry.ResolvedDestination)
	}
	return nil
}

func (fedex *FullEntryDetailsExtra) UnmarshalData(entry *tapApi.MizuEntry) error {
	var root tapApi.RequestResponsePair
	err := json.Unmarshal([]byte(entry.Entry), &root)
	if err != nil {
		return err
	}
	requestPayload := root.Request.Payload.(map[string]interface{})
	responsePayload := root.Response.Payload.(map[string]interface{})
	totalTime := root.Response.CaptureTime.Sub(root.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	fedex.Entry = NewEntry(root.Request.CaptureTime, requestPayload["details"].(interface{}).(*har.Request), responsePayload["details"].(interface{}).(*har.Response), totalTime)


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
	var root tapApi.RequestResponsePair
	err := json.Unmarshal([]byte(entry.Entry), &root)
	if err != nil {
		return err
	}
	requestPayload := root.Request.Payload.(map[string]interface{})
	responsePayload := root.Response.Payload.(map[string]interface{})
	totalTime := root.Response.CaptureTime.Sub(root.Request.CaptureTime).Round(time.Millisecond).Milliseconds()
	fewp.Entry = NewEntry(root.Request.CaptureTime, requestPayload["details"].(interface{}).(*har.Request), responsePayload["details"].(interface{}).(*har.Response), totalTime)


	_, resultPolicyToSend := rules.MatchRequestPolicy(fewp.Entry, entry.Service)
	fewp.RulesMatched = resultPolicyToSend
	fewp.Service = entry.Service
	return nil
}

//func RunValidationRulesState(harEntry har.Entry, service string) tapApi.ApplicableRules {
//	numberOfRules, resultPolicyToSend := rules.MatchRequestPolicy(harEntry, service)
//	statusPolicyToSend, latency, numberOfRules := rules.PassedValidationRules(resultPolicyToSend, numberOfRules)
//	ar := NewApplicableRules(statusPolicyToSend, latency, numberOfRules)
//	return ar
//}
