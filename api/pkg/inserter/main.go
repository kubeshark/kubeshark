package inserter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/antoniodipinto/ikisocket"
	"github.com/google/martian/har"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

func StartReadingEntries(harChannel chan *har.Entry, workingDir *string) {
	if workingDir != nil && *workingDir != "" {
		startReadingFiles(*workingDir)
	} else {
		startReadingChannel(harChannel)
	}
}

func startReadingFiles(workingDir string) {
	err := os.MkdirAll(workingDir, os.ModePerm)
	utils.CheckErr(err)

	for true {
		dir, _ := os.Open(workingDir)
		dirFiles, _ := dir.Readdir(-1)
		sort.Sort(utils.ByModTime(dirFiles))

		if len(dirFiles) == 0{
			fmt.Printf("Waiting for new files\n")
			time.Sleep(3 * time.Second)
			continue
		}
		fileInfo := dirFiles[0]
		inputFilePath := path.Join(workingDir, fileInfo.Name())
		file, err := os.Open(inputFilePath)
		utils.CheckErr(err)

		var inputHar har.HAR
		decErr := json.NewDecoder(bufio.NewReader(file)).Decode(&inputHar)
		utils.CheckErr(decErr)

		for _, entry := range inputHar.Log.Entries {
			time.Sleep(time.Millisecond * 250)
			saveHarToDb(*entry)
		}
		rmErr := os.Remove(inputFilePath)
		utils.CheckErr(rmErr)
	}
}

func startReadingChannel(harChannel chan *har.Entry) {
	for entry := range harChannel {
		saveHarToDb(*entry)
	}
}

func saveHarToDb(entry har.Entry) {
	entryBytes, _ := json.Marshal(entry)
	serviceName, urlPath := getServiceNameFromUrl(entry.Request.URL)
	source := getSourceNameFromHeaders(entry.Request.Headers)
	entryId := primitive.NewObjectID().Hex()
	mizuEntry := models.MizuEntry{
		EntryId:   entryId,
		Entry:     string(entryBytes), // simple way to store it and not convert to bytes
		Service:   serviceName,
		Url:       entry.Request.URL,
		Path:      urlPath,
		Method:    entry.Request.Method,
		Status:    entry.Response.Status,
		Source:    source,
		Timestamp: entry.StartedDateTime.UnixNano() / int64(time.Millisecond),
	}
	database.GetEntriesTable().Create(&mizuEntry)

	baseEntry := &models.BaseEntryDetails{
		Id:         entryId,
		Url:        entry.Request.URL,
		Service:    serviceName,
		Path:       urlPath,
		StatusCode: entry.Response.Status,
		Method:     entry.Request.Method,
		Source: 	source,
		Timestamp:  entry.StartedDateTime.UnixNano() / int64(time.Millisecond),
	}
	baseEntryBytes, _ := json.Marshal(&baseEntry)
	ikisocket.Broadcast(baseEntryBytes)

}

func getSourceNameFromHeaders(requestHeaders []har.Header) string{
	for _, header := range requestHeaders {
		if strings.ToLower(header.Name) == "x-up9-source" {
			return header.Value
		}
	}
	return ""
}
func getServiceNameFromUrl(inputUrl string) (string, string) {
	parsed, err := url.Parse(inputUrl)
	utils.CheckErr(err)
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host), parsed.Path
}
