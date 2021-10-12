package models

import (
	"encoding/json"

	tapApi "github.com/up9inc/mizu/tap/api"

	"mizuserver/pkg/rules"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
)

func GetEntry(r *tapApi.MizuEntry, v tapApi.DataUnmarshaler) error {
	return v.UnmarshalData(r)
}

type EntriesFilter struct {
	Limit     int    `form:"limit" validate:"required,min=1,max=200"`
	Operator  string `form:"operator" validate:"required,oneof='lt' 'gt'"`
	Timestamp int64  `form:"timestamp" validate:"required,min=1"`
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

type AuthStatus struct {
	Email string `json:"email"`
	Model string `json:"model"`
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

func RunValidationRulesState(harEntry har.Entry, service string) (tapApi.ApplicableRules, []rules.RulesMatched, bool) {
	resultPolicyToSend, isEnabled := rules.MatchRequestPolicy(harEntry, service)
	statusPolicyToSend, latency, numberOfRules := rules.PassedValidationRules(resultPolicyToSend)
	return tapApi.ApplicableRules{Status: statusPolicyToSend, Latency: latency, NumberOfRules: numberOfRules}, resultPolicyToSend, isEnabled
}
