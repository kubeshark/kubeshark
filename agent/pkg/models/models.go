package models

import (
	"encoding/json"
	"mizuserver/pkg/rules"
	"mizuserver/pkg/utils"

	tapApi "github.com/up9inc/mizu/tap/api"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
)

func GetEntry(r *tapApi.MizuEntry, v tapApi.DataUnmarshaler) error {
	return v.UnmarshalData(r)
}

// TODO: until we fixed the Rules feature
//func NewApplicableRules(status bool, latency int64, number int) tapApi.ApplicableRules {
//	ar := tapApi.ApplicableRules{}
//	ar.Status = status
//	ar.Latency = latency
//	ar.NumberOfRules = number
//	return ar
//}

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
	Data map[string]interface{} `json:"data,omitempty"`
}

type WebSocketTappedEntryMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tapApi.OutputChannelItem
}

type WebsocketOutboundLinkMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tap.OutboundLink
}

func CreateBaseEntryWebSocketMessage(base map[string]interface{}) ([]byte, error) {
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
	var pair tapApi.RequestResponsePair
	if err := json.Unmarshal([]byte(entry.Entry), &pair); err != nil {
		return err
	}
	harEntry, err := utils.NewEntry(&pair)
	if err != nil {
		return err
	}
	fewp.Entry = *harEntry

	_, resultPolicyToSend := rules.MatchRequestPolicy(fewp.Entry, entry.Service)
	fewp.RulesMatched = resultPolicyToSend
	fewp.Service = entry.Service
	return nil
}

// TODO: until we fixed the Rules feature
//func RunValidationRulesState(harEntry har.Entry, service string) tapApi.ApplicableRules {
//	numberOfRules, resultPolicyToSend := rules.MatchRequestPolicy(harEntry, service)
//	statusPolicyToSend, latency, numberOfRules := rules.PassedValidationRules(resultPolicyToSend, numberOfRules)
//	ar := NewApplicableRules(statusPolicyToSend, latency, numberOfRules)
//	return ar
//}
