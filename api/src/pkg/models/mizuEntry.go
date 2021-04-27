package models

import "time"

type MizuEntry struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Entry     string `json:"entry,omitempty" gorm:"column:entry"`
	EntryId   string `json:"entryId" gorm:"column:entryId"`
	Url       string `json:"url" gorm:"column:url"`
	Method    string `json:"method" gorm:"column:method"`
	Status    int    `json:"status" gorm:"column:status"`
	Source    string `json:"source" gorm:"column:source"`
	Service   string `json:"service" gorm:"column:service"`
	Timestamp int64  `json:"timestamp" gorm:"column:timestamp"`
	Path      string `json:"path" gorm:"column:path"`
}

type BaseEntryDetails struct {
	Id         string `json:"id,omitempty"`
	Url        string `json:"url,omitempty"`
	Service    string `json:"service,omitempty"`
	Path       string `json:"path,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
	Method     string `json:"method,omitempty"`
	Timestamp  int64  `json:"timestamp,omitempty"`
}

type EntryData struct {
	Entry string `json:"entry,omitempty"`
}
