package models

import (
	"encoding/json"
	"mizuserver/pkg/rules"

	tapApi "github.com/up9inc/mizu/tap/api"

	"github.com/google/martian/har"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
)

func GetEntry(r *tapApi.MizuEntry, v tapApi.DataUnmarshaler) error {
	return v.UnmarshalData(r)
}

type StandaloneConfig struct {
	TargetNamespaces []string `json:"targetNamespaces"`
	PodRegex         string   `json:"podRegex"`
}

type EntriesRequest struct {
	LeftOff   int    `form:"leftOff" validate:"required,min=-1"`
	Direction int    `form:"direction" validate:"required,oneof='1' '-1'"`
	Query     string `form:"query"`
	Limit     int    `form:"limit" validate:"required,min=1"`
	TimeoutMs int    `form:"timeoutMs" validate:"min=1"`
}

type EntriesResponse struct {
	Data []interface{}      `json:"data"`
	Meta *basenine.Metadata `json:"meta"`
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

type AuthStatus struct {
	Email string `json:"email"`
	Model string `json:"model"`
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
