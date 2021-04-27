package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/martian/har"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mizuserver/src/pkg/database"
	"mizuserver/src/pkg/models"
	"mizuserver/src/pkg/utils"
	"net/url"
	"os"
	"path"
	"sort"
)

func TestHarSavingFromFolder(inputDir string) {
	dir, _ := os.Open(inputDir)
	dirFiles, _ := dir.Readdir(-1)
	sort.Sort(utils.ByModTime(dirFiles))

	for _, fileInfo := range dirFiles {
		inputFilePath := path.Join(inputDir, fileInfo.Name())
		file, err := os.Open(inputFilePath)
		utils.CheckErr(err)

		var inputHar har.HAR
		decErr := json.NewDecoder(bufio.NewReader(file)).Decode(&inputHar)
		utils.CheckErr(decErr)

		for _, entry := range inputHar.Log.Entries {
			SaveHarToDb(*entry, "source")
		}
	}
}

func SaveHarToDb(entry har.Entry, source string) {
	entryBytes, _ := json.Marshal(entry)
	serviceName, urlPath := getServiceNameFromUrl(entry.Request.URL)
	mizuEntry := models.MizuEntry{
		EntryId:   primitive.NewObjectID().Hex(),
		Entry:     string(entryBytes), // simple way to store it and not convert to bytes
		Service:   serviceName,
		Url:       entry.Request.URL,
		Path:      urlPath,
		Method:    entry.Request.Method,
		Status:    entry.Response.Status,
		Source:    source,
		Timestamp: entry.StartedDateTime.Unix(),
	}
	database.GetEntriesTable().Create(&mizuEntry)
}

func getServiceNameFromUrl(inputUrl string) (string, string) {
	parsed, err := url.Parse(inputUrl)
	utils.CheckErr(err)
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host), parsed.Path
}
