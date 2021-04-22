package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/martian/har"
	"github.com/jinzhu/gorm"
	// "gorm.io/driver/sqlite"
	 _ "github.com/jinzhu/gorm/dialects/sqlite"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	FILEPATH := "/Users/roeegadot/Downloads/testing.har"
	file, err := os.Open(FILEPATH)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	reader := bufio.NewReader(file)
	dec := json.NewDecoder(reader)
	var inputHar har.HAR
	if err := dec.Decode(&inputHar); err != nil {
		fmt.Print(err)
		os.Exit(1)

	}

	for _, entry := range inputHar.Log.Entries {
		saveHarToDb(*entry)
	}

}

func getGormDB() *gorm.DB {
	db, _ := gorm.Open("sqlite", "roee.db")
	return db
}

func saveHarToDb(entry har.Entry) {
	entryData := entry
	mizuEntry := MizuEntry{
		EntryId: NewObjectID(),
		Entry: entryData,
		Url: entryData.Request.URL,
		Method: entryData.Request.Method,
		Status: entryData.Response.Status,
		Source: "",
		Service: "MyService",
	}

	if err := mizuEntry.Create(getGormDB()); err != nil {
		panic("cannot create")
	}
}

