package models

import (
	"encoding/json"

	"mizuserver/pkg/rules"
	"mizuserver/pkg/utils"
	"time"

	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
)

type DataUnmarshaler interface {
	UnmarshalData(*MizuEntry) error
}

func GetEntry(r *MizuEntry, v DataUnmarshaler) error {
	return v.UnmarshalData(r)
}

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
	EstimatedSizeBytes  int    `json:"-" gorm:"column:estimatedSizeBytes"`
}

type BaseEntryDetails struct {
	Id              string          `json:"id,omitempty"`
	Url             string          `json:"url,omitempty"`
	RequestSenderIp string          `json:"requestSenderIp,omitempty"`
	Service         string          `json:"service,omitempty"`
	Path            string          `json:"path,omitempty"`
	StatusCode      int             `json:"statusCode,omitempty"`
	Method          string          `json:"method,omitempty"`
	Timestamp       int64           `json:"timestamp,omitempty"`
	IsOutgoing      bool            `json:"isOutgoing,omitempty"`
	Latency         int64           `json:"latency,omitempty"`
	Rules           ApplicableRules `json:"rules,omitempty"`
}

type ApplicableRules struct {
	Latency       int64 `json:"latency,omitempty"`
	Status        bool  `json:"status,omitempty"`
	NumberOfRules int   `json:"numberOfRules,omitempty"`
}

type FullEntryDetails struct {
	har.Entry
}

type FullEntryDetailsExtra struct {
	har.Entry
}

func (bed *BaseEntryDetails) UnmarshalData(entry *MizuEntry) error {
	entryUrl := entry.Url
	service := entry.Service
	if entry.ResolvedDestination != "" {
		entryUrl = utils.SetHostname(entryUrl, entry.ResolvedDestination)
		service = utils.SetHostname(service, entry.ResolvedDestination)
	}
	bed.Id = entry.EntryId
	bed.Url = entryUrl
	bed.Service = service
	bed.Path = entry.Path
	bed.StatusCode = entry.Status
	bed.Method = entry.Method
	bed.Timestamp = entry.Timestamp
	bed.RequestSenderIp = entry.RequestSenderIp
	bed.IsOutgoing = entry.IsOutgoing
	return nil
}

func (fed *FullEntryDetails) UnmarshalData(entry *MizuEntry) error {
	if err := json.Unmarshal([]byte(entry.Entry), &fed.Entry); err != nil {
		return err
	}

	if entry.ResolvedDestination != "" {
		fed.Entry.Request.URL = utils.SetHostname(fed.Entry.Request.URL, entry.ResolvedDestination)
	}
	return nil
}

func (fedex *FullEntryDetailsExtra) UnmarshalData(entry *MizuEntry) error {
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
	Data *BaseEntryDetails `json:"data,omitempty"`
}

type WebSocketTappedEntryMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tap.OutputChannelItem
}

type WebsocketOutboundLinkMessage struct {
	*shared.WebSocketMessageMetadata
	Data *tap.OutboundLink
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

func (fewp *FullEntryWithPolicy) UnmarshalData(entry *MizuEntry) error {
	if err := json.Unmarshal([]byte(entry.Entry), &fewp.Entry); err != nil {
		return err
	}

	_, resultPolicyToSend := rules.MatchRequestPolicy(fewp.Entry, entry.Service)
	fewp.RulesMatched = resultPolicyToSend
	fewp.Service = entry.Service
	return nil
}

func RunValidationRulesState(harEntry har.Entry, service string) ApplicableRules {
	_, resultPolicyToSend := rules.MatchRequestPolicy(harEntry, service)
	return rules.PassedValidationRules(resultPolicyToSend)
}
