package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"mizuserver/pkg/database"
	"mizuserver/pkg/holder"
	"mizuserver/pkg/providers"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/google/martian/har"
	"github.com/romana/rlog"
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

func StartReadingEntries(harChannel <-chan *tapApi.OutputChannelItem, workingDir *string, extensionsMap map[string]*tapApi.Extension) {
	if workingDir != nil && *workingDir != "" {
		startReadingFiles(*workingDir)
	} else {
		startReadingChannel(harChannel, extensionsMap)
	}
}

func startReadingFiles(workingDir string) {
	if err := os.MkdirAll(workingDir, os.ModePerm); err != nil {
		rlog.Errorf("Failed to make dir: %s, err: %v", workingDir, err)
		return
	}

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

		rmErr := os.Remove(inputFilePath)
		utils.CheckErr(rmErr)
	}
}

func startReadingChannel(outputItems <-chan *tapApi.OutputChannelItem, extensionsMap map[string]*tapApi.Extension) {
	if outputItems == nil {
		panic("Channel of captured messages is nil")
	}

	disableOASValidation := false
	ctx, doc, contractContent, router, err := loadOAS()
	if err != nil {
		rlog.Infof("Disabled OAS validation: %s\n", err.Error())
		disableOASValidation = true
	}

	for item := range outputItems {
		providers.EntryAdded()

		extension := extensionsMap[item.Protocol.Name]
		resolvedSource, resolvedDestionation := resolveIP(item.ConnectionInfo)
		mizuEntry := extension.Dissector.Analyze(item, primitive.NewObjectID().Hex(), resolvedSource, resolvedDestionation)
		baseEntry := extension.Dissector.Summarize(mizuEntry)
		mizuEntry.EstimatedSizeBytes = getEstimatedEntrySizeBytes(mizuEntry)
		if extension.Protocol.Name == "http" {
			if !disableOASValidation {
				var httpPair tapApi.HTTPRequestResponsePair
				json.Unmarshal([]byte(mizuEntry.Entry), &httpPair)

				contract := handleOAS(ctx, doc, router, httpPair.Request.Payload.RawRequest, httpPair.Response.Payload.RawResponse, contractContent)
				baseEntry.ContractStatus = contract.Status
				mizuEntry.ContractStatus = contract.Status
				mizuEntry.ContractRequestReason = contract.RequestReason
				mizuEntry.ContractResponseReason = contract.ResponseReason
				mizuEntry.ContractContent = contract.Content
			}

			var pair tapApi.RequestResponsePair
			json.Unmarshal([]byte(mizuEntry.Entry), &pair)
			harEntry, err := utils.NewEntry(&pair)
			if err == nil {
				rules, _, _ := models.RunValidationRulesState(*harEntry, mizuEntry.Service)
				baseEntry.Rules = rules
			}
		}
		database.CreateEntry(mizuEntry)

		baseEntryBytes, _ := models.CreateBaseEntryWebSocketMessage(baseEntry)
		BroadcastToBrowserClients(baseEntryBytes)
	}
}

func resolveIP(connectionInfo *tapApi.ConnectionInfo) (resolvedSource string, resolvedDestination string) {
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
	return resolvedSource, resolvedDestination
}

func CheckIsServiceIP(address string) bool {
	if k8sResolver == nil {
		return false
	}
	return k8sResolver.CheckIsServiceIP(address)
}

// gives a rough estimate of the size this will take up in the db, good enough for maintaining db size limit accurately
func getEstimatedEntrySizeBytes(mizuEntry *tapApi.MizuEntry) int {
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
