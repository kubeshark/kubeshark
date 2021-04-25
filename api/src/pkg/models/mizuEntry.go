package models

import (
	"gorm.io/gorm"
)

type MizuEntry struct {
	gorm.Model
	Entry       string `json:"entry,omitempty" gorm:"column:entry"`
	EntryId     string `json:"entryId" gorm:"column:entryId"`
	Url         string `json:"url" gorm:"column:url"`
	Method      string `json:"method" gorm:"column:method"`
	Status      int    `json:"status" gorm:"column:status"`
	Source      string `json:"source" gorm:"column:source"`
	ServiceName string `json:"serviceName" gorm:"column:serviceName"`
	Timestamp   int64  `json:"timestamp" gorm:"column:timestamp"`
	Path        string `json:"path" gorm:"column:path"`
}

type BaseEntryDetails struct {
	Id          string `json:"id,omitempty"`
	EntryId     string `json:"entryId,omitempty"`
	Url         string `json:"url,omitempty"`
	ServiceName string `json:"serviceName,omitempty"`
	Path        string `json:"path,omitempty"`
	StatusCode  int    `json:"statusCode,omitempty"`
	Method      string `json:"method,omitempty"`
	Timestamp   int64  `json:"timestamp,omitempty"`
}


type EntryData struct {
	Entry string `json:"entry,omitempty"`
}