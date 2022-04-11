package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/agent/pkg/models"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/up9inc/mizu/agent/pkg/dependency"
	"github.com/up9inc/mizu/agent/pkg/elastic"
	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/agent/pkg/holder"
	"github.com/up9inc/mizu/agent/pkg/providers"

	"github.com/up9inc/mizu/agent/pkg/servicemap"

	"github.com/up9inc/mizu/agent/pkg/resolver"
	"github.com/up9inc/mizu/agent/pkg/utils"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"

	basenine "github.com/up9inc/basenine/client/go"
)

var k8sResolver *resolver.Resolver

func StartResolving(namespace string) {
	errOut := make(chan error, 100)
	res, err := resolver.NewFromInCluster(errOut, namespace)
	if err != nil {
		logger.Log.Infof("error creating k8s resolver %s", err)
		return
	}
	ctx := context.Background()
	res.Start(ctx)
	go func() {
		for {
			err := <-errOut
			logger.Log.Infof("name resolving error %s", err)
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
		logger.Log.Errorf("Failed to make dir: %s, err: %v", workingDir, err)
		return
	}

	for {
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
			logger.Log.Infof("Waiting for new files")
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

	connection, err := basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
	if err != nil {
		panic(err)
	}
	connection.InsertMode()

	disableOASValidation := false
	ctx := context.Background()
	doc, contractContent, router, err := loadOAS(ctx)
	if err != nil {
		logger.Log.Infof("Disabled OAS validation: %s", err.Error())
		disableOASValidation = true
	}

	for item := range outputItems {
		extension := extensionsMap[item.Protocol.Name]
		resolvedSource, resolvedDestionation, namespace := resolveIP(item.ConnectionInfo)
		mizuEntry := extension.Dissector.Analyze(item, resolvedSource, resolvedDestionation, namespace)
		if extension.Protocol.Name == "http" {
			if !disableOASValidation {
				var httpPair tapApi.HTTPRequestResponsePair
				if err := json.Unmarshal([]byte(mizuEntry.HTTPPair), &httpPair); err != nil {
					logger.Log.Error(err)
				}

				contract := handleOAS(ctx, doc, router, httpPair.Request.Payload.RawRequest, httpPair.Response.Payload.RawResponse, contractContent)
				mizuEntry.ContractStatus = contract.Status
				mizuEntry.ContractRequestReason = contract.RequestReason
				mizuEntry.ContractResponseReason = contract.ResponseReason
				mizuEntry.ContractContent = contract.Content
			}

			harEntry, err := har.NewEntry(mizuEntry.Request, mizuEntry.Response, mizuEntry.StartTime, mizuEntry.ElapsedTime)
			if err == nil {
				rules, _, _ := models.RunValidationRulesState(*harEntry, mizuEntry.Destination.Name)
				mizuEntry.Rules = rules
			}
		}

		data, err := json.Marshal(mizuEntry)
		if err != nil {
			panic(err)
		}

		providers.EntryAdded(len(data))

		connection.SendText(string(data))

		serviceMapGenerator := dependency.GetInstance(dependency.ServiceMapGeneratorDependency).(servicemap.ServiceMapSink)
		serviceMapGenerator.NewTCPEntry(mizuEntry.Source, mizuEntry.Destination, &item.Protocol)

		elastic.GetInstance().PushEntry(mizuEntry)
	}
}

func resolveIP(connectionInfo *tapApi.ConnectionInfo) (resolvedSource string, resolvedDestination string, namespace string) {
	if k8sResolver != nil {
		unresolvedSource := connectionInfo.ClientIP
		resolvedSourceObject := k8sResolver.Resolve(unresolvedSource)
		if resolvedSourceObject == nil {
			logger.Log.Debugf("Cannot find resolved name to source: %s", unresolvedSource)
			if os.Getenv("SKIP_NOT_RESOLVED_SOURCE") == "1" {
				return
			}
		} else {
			resolvedSource = resolvedSourceObject.FullAddress
			namespace = resolvedSourceObject.Namespace
		}

		unresolvedDestination := fmt.Sprintf("%s:%s", connectionInfo.ServerIP, connectionInfo.ServerPort)
		resolvedDestinationObject := k8sResolver.Resolve(unresolvedDestination)
		if resolvedDestinationObject == nil {
			logger.Log.Debugf("Cannot find resolved name to dest: %s", unresolvedDestination)
			if os.Getenv("SKIP_NOT_RESOLVED_DEST") == "1" {
				return
			}
		} else {
			resolvedDestination = resolvedDestinationObject.FullAddress
			// Overwrite namespace (if it was set according to the source)
			// Only overwrite if non-empty
			if resolvedDestinationObject.Namespace != "" {
				namespace = resolvedDestinationObject.Namespace
			}
		}
	}
	return resolvedSource, resolvedDestination, namespace
}

func CheckIsServiceIP(address string) bool {
	if k8sResolver == nil {
		return false
	}
	return k8sResolver.CheckIsServiceIP(address)
}
