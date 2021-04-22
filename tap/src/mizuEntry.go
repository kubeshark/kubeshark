package main

import (
	"gorm.io/gorm"
)

type MizuEntry struct {
	gorm.Model
	Entry   string `json:"entry,omitempty"`
	EntryId string `json:"entryId"`
	Url     string `json:"url"`
	Method  string `json:"method"`
	Status  int    `json:"status"`
	Source  string `json:"source"`
	Service string `json:"serviceName"`
}
