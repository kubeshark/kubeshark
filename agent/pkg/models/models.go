package models

import (
	"encoding/json"

	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/agent/pkg/rules"
	tapApi "github.com/up9inc/mizu/tap/api"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
)

type EntriesRequest struct {
	LeftOff   string `form:"leftOff" validate:"required"`
	Direction int    `form:"direction" validate:"required,oneof='1' '-1'"`
	Query     string `form:"query"`
	Limit     int    `form:"limit" validate:"required,min=1"`
	TimeoutMs int    `form:"timeoutMs" validate:"min=1"`
}

type SingleEntryRequest struct {
	Query string `form:"query"`
}

type EntriesResponse struct {
	Data []interface{}      `json:"data"`
	Meta *basenine.Metadata `json:"meta"`
}

type WebSocketEntryMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tapApi.BaseEntry `json:"data,omitempty"`
}

type WebSocketFullEntryMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tapApi.Entry `json:"data,omitempty"`
}

type WebSocketTappedEntryMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tapApi.OutputChannelItem
}

type ToastMessage struct {
	Type      string `json:"type"`
	AutoClose uint   `json:"autoClose"`
	Text      string `json:"text"`
}

type WebSocketToastMessage struct {
	*shared.WebSocketMessageMetadata
	Data *ToastMessage `json:"data,omitempty"`
}

type WebSocketQueryMetadataMessage struct {
	*shared.WebSocketMessageMetadata
	Data *basenine.Metadata `json:"data,omitempty"`
}

type WebSocketStartTimeMessage struct {
	*shared.WebSocketMessageMetadata
	Data int64 `json:"data"`
}

func CreateBaseEntryWebSocketMessage(base *tapApi.BaseEntry) ([]byte, error) {
	message := &WebSocketEntryMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebSocketMessageTypeEntry,
		},
		Data: base,
	}
	return json.Marshal(message)
}

func CreateFullEntryWebSocketMessage(entry *tapApi.Entry) ([]byte, error) {
	message := &WebSocketFullEntryMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebSocketMessageTypeFullEntry,
		},
		Data: entry,
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

func CreateWebsocketToastMessage(base *ToastMessage) ([]byte, error) {
	message := &WebSocketToastMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebSocketMessageTypeToast,
		},
		Data: base,
	}
	return json.Marshal(message)
}

func CreateWebsocketQueryMetadataMessage(base *basenine.Metadata) ([]byte, error) {
	message := &WebSocketQueryMetadataMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebSocketMessageTypeQueryMetadata,
		},
		Data: base,
	}
	return json.Marshal(message)
}

func CreateWebsocketStartTimeMessage(base int64) ([]byte, error) {
	message := &WebSocketStartTimeMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebSocketMessageTypeStartTime,
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
