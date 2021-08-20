package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"mizuserver/pkg/database"
	"mizuserver/pkg/holder"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/google/martian/har"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/tap"
	tapApi "github.com/up9inc/mizu/tap/api"

	"mizuserver/pkg/models"
	"mizuserver/pkg/resolver"
	"mizuserver/pkg/utils"
)

var k8sResolver *resolver.Resolver

func StartResolving(namespace string) {
	errOut := make(chan error, 100)
	res, err := resolver.NewFromInCluster(errOut, namespace)
	if err != nil {
		rlog.Infof("error creating k8s resolver %s", err)
		return
	}
	ctx := context.Background()
	res.Start(ctx)
	go func() {
		for {
			select {
			case err := <-errOut:
				rlog.Infof("name resolving error %s", err)
			}
		}
	}()

	k8sResolver = res
	holder.SetResolver(res)
}

func StartReadingEntries(harChannel <-chan *tapApi.OutputChannelItem, workingDir *string) {
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

		var harFiles []os.FileInfo
		for _, fileInfo := range dirFiles {
			if strings.HasSuffix(fileInfo.Name(), ".har") {
				harFiles = append(harFiles, fileInfo)
			}
		}
		sort.Sort(utils.ByModTime(harFiles))

		if len(harFiles) == 0 {
			rlog.Infof("Waiting for new files\n")
			time.Sleep(3 * time.Second)
			continue
		}
		fileInfo := harFiles[0]
		inputFilePath := path.Join(workingDir, fileInfo.Name())
		file, err := os.Open(inputFilePath)
		utils.CheckErr(err)

		var inputHar har.HAR
		decErr := json.NewDecoder(bufio.NewReader(file)).Decode(&inputHar)
		utils.CheckErr(decErr)

		// for _, entry := range inputHar.Log.Entries {
		// 	time.Sleep(time.Millisecond * 250)
		// 	// connectionInfo := &tap.ConnectionInfo{
		// 	// 	ClientIP:   fileInfo.Name(),
		// 	// 	ClientPort: "",
		// 	// 	ServerIP:   "",
		// 	// 	ServerPort: "",
		// 	// 	IsOutgoing: false,
		// 	// }
		// 	// saveHarToDb(entry, connectionInfo)
		// }
		rmErr := os.Remove(inputFilePath)
		utils.CheckErr(rmErr)
	}
}

func startReadingChannel(outputItems <-chan *tapApi.OutputChannelItem) {
	if outputItems == nil {
		panic("Channel of captured messages is nil")
	}

	for item := range outputItems {
		fmt.Printf("item: %+v\n", item)
		var req *http.Request
		marshedReq, _ := json.Marshal(item.Data.Request.Orig)
		json.Unmarshal(marshedReq, &req)
		var res *http.Response
		marshedRes, _ := json.Marshal(item.Data.Response.Orig)
		json.Unmarshal(marshedRes, &res)
		// NOTE: With this call, the incoming data is sent to the last WebSocket (that the web UI communicates).
		if harEntry, err := models.NewEntry(req, item.Data.Request.CaptureTime, res, item.Data.Response.CaptureTime); err == nil {
			saveHarToDb(harEntry, item.ConnectionInfo)
		} else {
			rlog.Errorf("Error when creating HTTP entry")
		}
	}
}

func StartReadingOutbound(outboundLinkChannel <-chan *tap.OutboundLink) {
	// tcpStreamFactory will block on write to channel. Empty channel to unblock.
	// TODO: Make write to channel optional.
	for range outboundLinkChannel {
	}
}

func saveHarToDb(entry *har.Entry, connectionInfo *tapApi.ConnectionInfo) {
	entryBytes, _ := json.Marshal(entry)
	serviceName, urlPath := getServiceNameFromUrl(entry.Request.URL)
	entryId := primitive.NewObjectID().Hex()
	var (
		resolvedSource      string
		resolvedDestination string
	)
	if k8sResolver != nil {
		unresolvedSource := connectionInfo.ClientIP
		resolvedSource = k8sResolver.Resolve(unresolvedSource)
		if resolvedSource == "" {
			rlog.Debugf("Cannot find resolved name to source: %s\n", unresolvedSource)
			if os.Getenv("SKIP_NOT_RESOLVED_SOURCE") == "1" {
				return
			}
		}
		unresolvedDestination := fmt.Sprintf("%s:%s", connectionInfo.ServerIP, connectionInfo.ServerPort)
		resolvedDestination = k8sResolver.Resolve(unresolvedDestination)
		if resolvedDestination == "" {
			rlog.Debugf("Cannot find resolved name to dest: %s\n", unresolvedDestination)
			if os.Getenv("SKIP_NOT_RESOLVED_DEST") == "1" {
				return
			}
		}
	}

	mizuEntry := models.MizuEntry{
		EntryId:             entryId,
		Entry:               string(entryBytes), // simple way to store it and not convert to bytes
		Service:             serviceName,
		Url:                 entry.Request.URL,
		Path:                urlPath,
		Method:              entry.Request.Method,
		Status:              entry.Response.Status,
		RequestSenderIp:     connectionInfo.ClientIP,
		Timestamp:           entry.StartedDateTime.UnixNano() / int64(time.Millisecond),
		ResolvedSource:      resolvedSource,
		ResolvedDestination: resolvedDestination,
		IsOutgoing:          connectionInfo.IsOutgoing,
	}
	mizuEntry.EstimatedSizeBytes = getEstimatedEntrySizeBytes(mizuEntry)
	database.CreateEntry(&mizuEntry)

	baseEntry := models.BaseEntryDetails{}
	if err := models.GetEntry(&mizuEntry, &baseEntry); err != nil {
		return
	}
	baseEntry.Rules = models.RunValidationRulesState(*entry, serviceName)
	baseEntry.Latency = entry.Timings.Receive
	baseEntryBytes, _ := models.CreateBaseEntryWebSocketMessage(&baseEntry)
	BroadcastToBrowserClients(baseEntryBytes)
}

func getServiceNameFromUrl(inputUrl string) (string, string) {
	parsed, err := url.Parse(inputUrl)
	utils.CheckErr(err)
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host), parsed.Path
}

func CheckIsServiceIP(address string) bool {
	return k8sResolver.CheckIsServiceIP(address)
}

// gives a rough estimate of the size this will take up in the db, good enough for maintaining db size limit accurately
func getEstimatedEntrySizeBytes(mizuEntry models.MizuEntry) int {
	sizeBytes := len(mizuEntry.Entry)
	sizeBytes += len(mizuEntry.EntryId)
	sizeBytes += len(mizuEntry.Service)
	sizeBytes += len(mizuEntry.Url)
	sizeBytes += len(mizuEntry.Method)
	sizeBytes += len(mizuEntry.RequestSenderIp)
	sizeBytes += len(mizuEntry.ResolvedDestination)
	sizeBytes += len(mizuEntry.ResolvedSource)
	sizeBytes += 8 // Status bytes (sqlite integer is always 8 bytes)
	sizeBytes += 8 // Timestamp bytes
	sizeBytes += 8 // SizeBytes bytes
	sizeBytes += 1 // IsOutgoing bytes

	return sizeBytes
}
