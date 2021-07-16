package models

import (
	"encoding/json"
	"time"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
)

type MizuEntry struct {
	ID                  uint `gorm:"primarykey"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Entry               string `json:"entry,omitempty" gorm:"column:entry"`
	EntryId             string `json:"entryId" gorm:"column:entryId"`
	Url                 string `json:"url" gorm:"column:url"`
	Method              string `json:"method" gorm:"column:method"`
	Status              int    `json:"status" gorm:"column:status"`
	RequestSenderIp     string `json:"requestSenderIp" gorm:"column:requestSenderIp"`
	Service             string `json:"service" gorm:"column:service"`
	Timestamp           int64  `json:"timestamp" gorm:"column:timestamp"`
	Path                string `json:"path" gorm:"column:path"`
	ResolvedSource      string `json:"resolvedSource,omitempty" gorm:"column:resolvedSource"`
	ResolvedDestination string `json:"resolvedDestination,omitempty" gorm:"column:resolvedDestination"`
	IsOutgoing          bool   `json:"isOutgoing,omitempty" gorm:"column:isOutgoing"`
}

type BaseEntryDetails struct {
	Id              string `json:"id,omitempty"`
	Url             string `json:"url,omitempty"`
	RequestSenderIp string `json:"requestSenderIp,omitempty"`
	Service         string `json:"service,omitempty"`
	Path            string `json:"path,omitempty"`
	StatusCode      int    `json:"statusCode,omitempty"`
	Method          string `json:"method,omitempty"`
	Timestamp       int64  `json:"timestamp,omitempty"`
	IsOutgoing      bool   `json:"isOutgoing,omitempty"`
	ApplicableRules string `json:"applicableRules,omitempty"`
}

type EntryData struct {
	Entry               string `json:"entry,omitempty"`
	ResolvedDestination string `json:"resolvedDestination,omitempty" gorm:"column:resolvedDestination"`
	Service             string `json:"service,omitempty"`
}

type EntriesFilter struct {
	Limit     int    `query:"limit" validate:"required,min=1,max=200"`
	Operator  string `query:"operator" validate:"required,oneof='lt' 'gt'"`
	Timestamp int64  `query:"timestamp" validate:"required,min=1"`
}

type UploadEntriesRequestBody struct {
	Dest string `query:"dest"`
}

type HarFetchRequestBody struct {
	From int64 `query:"from"`
	To   int64 `query:"to"`
}

type WebSocketEntryMessage struct {
	*shared.WebSocketMessageMetadata
	Data *BaseEntryDetails `json:"data,omitempty"`
}

type WebSocketTappedEntryMessage struct {
	*shared.WebSocketMessageMetadata
	Data                  *tap.OutputChannelItem
	PassedValidationRules string
}

func CreateBaseEntryWebSocketMessage(base *BaseEntryDetails) ([]byte, error) {
	message := &WebSocketEntryMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebSocketMessageTypeEntry,
		},
		Data: base,
	}
	return json.Marshal(message)
}

func CreateWebsocketTappedEntryMessage(base *tap.OutputChannelItem) ([]byte, error) {
	message := &WebSocketTappedEntryMessage{
		WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
			MessageType: shared.WebSocketMessageTypeTappedEntry,
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
	Source string `json:"_source"`
}

type FullEntryWithPolicy struct {
	RulesMatched []shared.RulesMatched `json:"rulesMatched,omitempty"`
	Entry        har.Entry             `json:"entry"`
	Service      string                `json:"service"`
}

func RunValidationRulesState(fullEntry *har.Entry, service string) string {
	numberOfRules, resultPolicyToSend := shared.MatchRequestPolicy(*fullEntry, service)
	statusPolicyToSend := shared.PassedValidationRules(resultPolicyToSend, numberOfRules)
	return statusPolicyToSend
}
