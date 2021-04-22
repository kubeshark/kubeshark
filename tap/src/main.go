package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/martian/har"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)


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
	db, _ := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})

	migErr := db.AutoMigrate(&MizuEntry{})
	if migErr != nil {
		panic("Cannot run migration")
	}
	return db
}

func saveHarToDb(entry har.Entry) {
	a, _ := json.Marshal(entry)
	mizuEntry := MizuEntry{
		EntryId: NewObjectID().Hex(),
		// simple way to store it and not convert to bytes
		Entry: string(a),
		Url: (&entry).Request.URL,
		Method: (&entry).Request.Method,
		Status: (&entry).Response.Status,
		Source: "",
		Service: "MyService",
	}

	if err := mizuEntry.Create(getGormDB()); err != nil {
		fmt.Print(err)
		panic("cannot create")
	}
}


