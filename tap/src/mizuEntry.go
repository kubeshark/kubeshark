package main

import (
	"gorm.io/gorm"
)


//go:generate goqueryset -in mizuEntry.go
// gen:qs
type MizuEntry struct {
	gorm.Model
	// The Entry itself (as string)
	Entry string `json:"entry,omitempty"`
	//TODO: here we will add fields we need to query for

	EntryId string `json:"entryId"`
	Url string `json:"url"`
	Method string `json:"method"`
	Status int `json:"status"`
	Source string `json:"source"`
	Service string `json:"serviceName"`
}
