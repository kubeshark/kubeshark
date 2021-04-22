package main

import (
	"github.com/jinzhu/gorm"
	"github.com/google/martian/har"
)


//go:generate goqueryset -in mizuEntry.go
// gen:qs
type MizuEntry struct {
	gorm.Model
	// The Entry itself
	Entry har.Entry `json:"entry,omitempty"`
	//TODO: here we will add fields we need to query for

	EntryId ObjectID `json:"entryId"`
	Url string `json:"url"`
	Method string `json:"method"`
	Status int `json:"status"`
	Source string `json:"source"`
	Service string `json:"serviceName"`
}
