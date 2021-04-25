package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/martian/har"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"mizuserver/src/pkg/models"
	"os"
)

func testHarSavingToDBFromFile() {
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
		saveHarToDb(*entry, "service", "source")
	}

}

func getGormDB(databaseFilePath string) *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(databaseFilePath), &gorm.Config{})
	_ = db.AutoMigrate(&models.MizuEntry{}) // this will ensure table is created
	return db
}

func saveHarToDb(entry har.Entry, serviceName string, source string) {
	entryBytes, _ := json.Marshal(entry)
	mizuEntry := models.MizuEntry{
		EntryId: primitive.NewObjectID().Hex(),
		Entry:   string(entryBytes), // simple way to store it and not convert to bytes
		Url:     entry.Request.URL,
		Method:  entry.Request.Method,
		Status:  entry.Response.Status,
		Source:  source,
		Service: serviceName,
	}
	getGormDB("entries.db").Create(&mizuEntry)
}
