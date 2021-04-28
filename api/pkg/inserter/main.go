package inserter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/martian/har"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"io/fs"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
	"net/url"
	"os"
	"path"
	"sort"
	"time"
)


func IsEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true
	}
	return false // Either not empty or error, suits both cases
}

func StartReadingFiles(workingDir string) {
	err := os.MkdirAll(workingDir, fs.ModeDir)
	utils.CheckErr(err)

	for true {
		if IsEmpty(workingDir) {
			fmt.Printf("Waiting for new files\n")
			time.Sleep(5 * time.Second)
			continue
		}

		dir, _ := os.Open(workingDir)
		dirFiles, _ := dir.Readdir(-1)
		sort.Sort(utils.ByModTime(dirFiles))

		fileInfo := dirFiles[0]
		inputFilePath := path.Join(workingDir, fileInfo.Name())
		file, err := os.Open(inputFilePath)
		utils.CheckErr(err)

		var inputHar har.HAR
		decErr := json.NewDecoder(bufio.NewReader(file)).Decode(&inputHar)
		utils.CheckErr(decErr)

		for _, entry := range inputHar.Log.Entries 	{
			SaveHarToDb(*entry, "")
		}
		rmErr := os.Remove(inputFilePath)
		utils.CheckErr(rmErr)
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
