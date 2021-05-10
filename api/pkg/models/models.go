package models

import "time"

type MizuEntry struct {
	ID              uint `gorm:"primarykey"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Entry           string `json:"entry,omitempty" gorm:"column:entry"`
	EntryId         string `json:"entryId" gorm:"column:entryId"`
	Url             string `json:"url" gorm:"column:url"`
	Method          string `json:"method" gorm:"column:method"`
	Status          int    `json:"status" gorm:"column:status"`
	RequestSenderIp string `json:"requestSenderIp" gorm:"column:requestSenderIp"`
	Service         string `json:"service" gorm:"column:service"`
	Timestamp       int64  `json:"timestamp" gorm:"column:timestamp"`
	Path            string `json:"path" gorm:"column:path"`
	ResolvedSource *string `json:"resolvedSource,omitempty" gorm:"column:resolvedSource"`
	ResolvedDestination *string `json:"resolvedDestination,omitempty" gorm:"column:resolvedDestination"`
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
}

type EntryData struct {
	Entry string `json:"entry,omitempty"`
	ResolvedDestination *string `json:"resolvedDestination,omitempty" gorm:"column:resolvedDestination"`
}

type EntriesFilter struct {
	Limit     int    `query:"limit" validate:"required,min=1,max=200"`
	Operator  string `query:"operator" validate:"required,oneof='lt' 'gt'"`
	Timestamp int64  `query:"timestamp" validate:"required,min=1"`
}
