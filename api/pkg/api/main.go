package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/martian/har"
	"github.com/up9inc/mizu/tap"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/resolver"
	"mizuserver/pkg/utils"
	"net/url"
	"os"
	"path"
	"sort"
	"time"
)

var k8sResolver *resolver.Resolver

func init() {
	errOut := make(chan error, 100)
	res, err := resolver.NewFromInCluster(errOut)
	if err != nil {
		fmt.Printf("error creating k8s resolver %s", err)
		return
	}
	ctx := context.Background()
	res.Start(ctx)
	go func() {
		for {
			select {
			case err := <-errOut:
				fmt.Printf("name resolving error %s", err)
			}
		}
	}()

	k8sResolver = res
}

func StartReadingEntries(harChannel <-chan *tap.OutputChannelItem, workingDir *string) {
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

		if len(dirFiles) == 0 {
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
			saveHarToDb(entry, fileInfo.Name())
		}
		rmErr := os.Remove(inputFilePath)
		utils.CheckErr(rmErr)
	}
}

func startReadingChannel(outputItems <-chan *tap.OutputChannelItem) {
	if outputItems == nil {
		panic("Channel of captured messages is nil")
	}

	for item := range outputItems {
		saveHarToDb(item.HarEntry, item.RequestSenderIp)
	}
}

func saveHarToDb(entry *har.Entry, sender string) {
	entryBytes, _ := json.Marshal(entry)
	serviceName, urlPath, serviceHostName := getServiceNameFromUrl(entry.Request.URL)
	entryId := primitive.NewObjectID().Hex()
	var (
		resolvedSource      *string
		resolvedDestination *string
	)
	if k8sResolver != nil {
		resolvedSource = k8sResolver.Resolve(sender)
		resolvedDestination = k8sResolver.Resolve(serviceHostName)
	}
	mizuEntry := models.MizuEntry{
		EntryId:             entryId,
		Entry:               string(entryBytes), // simple way to store it and not convert to bytes
		Service:             serviceName,
		Url:                 entry.Request.URL,
		Path:                urlPath,
		Method:              entry.Request.Method,
		Status:              entry.Response.Status,
		RequestSenderIp:     sender,
		Timestamp:           entry.StartedDateTime.UnixNano() / int64(time.Millisecond),
		ResolvedSource:      resolvedSource,
		ResolvedDestination: resolvedDestination,
	}
	database.GetEntriesTable().Create(&mizuEntry)

	baseEntry := utils.GetResolvedBaseEntry(mizuEntry)
	baseEntryBytes, _ := models.CreateBaseEntryWebSocketMessage(&baseEntry)
	broadcastToBrowserClients(baseEntryBytes)
}

func getServiceNameFromUrl(inputUrl string) (string, string, string) {
	parsed, err := url.Parse(inputUrl)
	utils.CheckErr(err)
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host), parsed.Path, parsed.Host
}
