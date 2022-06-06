package app

import (
	"fmt"
	"sort"
	"time"

	"github.com/antelman107/net-wait-go/wait"
	"github.com/op/go-logging"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/utils"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/dbgctl"
	tapApi "github.com/up9inc/mizu/tap/api"
	amqpExt "github.com/up9inc/mizu/tap/extensions/amqp"
	httpExt "github.com/up9inc/mizu/tap/extensions/http"
	kafkaExt "github.com/up9inc/mizu/tap/extensions/kafka"
	redisExt "github.com/up9inc/mizu/tap/extensions/redis"
)

var (
	Extensions    []*tapApi.Extension          // global
	ExtensionsMap map[string]*tapApi.Extension // global
)

func LoadExtensions() {
	Extensions = make([]*tapApi.Extension, 0)
	ExtensionsMap = make(map[string]*tapApi.Extension)

	extensionHttp := &tapApi.Extension{}
	dissectorHttp := httpExt.NewDissector()
	dissectorHttp.Register(extensionHttp)
	extensionHttp.Dissector = dissectorHttp
	Extensions = append(Extensions, extensionHttp)
	ExtensionsMap[extensionHttp.Protocol.Name] = extensionHttp

	if !dbgctl.MizuTapperDisableNonHttpExtensions {
		extensionAmqp := &tapApi.Extension{}
		dissectorAmqp := amqpExt.NewDissector()
		dissectorAmqp.Register(extensionAmqp)
		extensionAmqp.Dissector = dissectorAmqp
		Extensions = append(Extensions, extensionAmqp)
		ExtensionsMap[extensionAmqp.Protocol.Name] = extensionAmqp

		extensionKafka := &tapApi.Extension{}
		dissectorKafka := kafkaExt.NewDissector()
		dissectorKafka.Register(extensionKafka)
		extensionKafka.Dissector = dissectorKafka
		Extensions = append(Extensions, extensionKafka)
		ExtensionsMap[extensionKafka.Protocol.Name] = extensionKafka

		extensionRedis := &tapApi.Extension{}
		dissectorRedis := redisExt.NewDissector()
		dissectorRedis.Register(extensionRedis)
		extensionRedis.Dissector = dissectorRedis
		Extensions = append(Extensions, extensionRedis)
		ExtensionsMap[extensionRedis.Protocol.Name] = extensionRedis
	}

	sort.Slice(Extensions, func(i, j int) bool {
		return Extensions[i].Protocol.Priority < Extensions[j].Protocol.Priority
	})

	api.InitExtensionsMap(ExtensionsMap)
}

func ConfigureBasenineServer(host string, port string, dbSize int64, logLevel logging.Level, insertionFilter string) {
	if !wait.New(
		wait.WithProto("tcp"),
		wait.WithWait(200*time.Millisecond),
		wait.WithBreak(50*time.Millisecond),
		wait.WithDeadline(20*time.Second),
		wait.WithDebug(logLevel == logging.DEBUG),
	).Do([]string{fmt.Sprintf("%s:%s", host, port)}) {
		logger.Log.Panicf("Basenine is not available!")
	}

	if err := basenine.Limit(host, port, dbSize); err != nil {
		logger.Log.Panicf("Error while limiting database size: %v", err)
	}

	// Define the macros
	for _, extension := range Extensions {
		macros := extension.Dissector.Macros()
		for macro, expanded := range macros {
			if err := basenine.Macro(host, port, macro, expanded); err != nil {
				logger.Log.Panicf("Error while adding a macro: %v", err)
			}
		}
	}

	// Set the insertion filter that comes from the config
	if err := basenine.InsertionFilter(host, port, insertionFilter); err != nil {
		logger.Log.Errorf("Error while setting the insertion filter: %v", err)
	}

	utils.StartTime = time.Now().UnixNano() / int64(time.Millisecond)
}

func GetEntryInputChannel() chan *tapApi.OutputChannelItem {
	outputItemsChannel := make(chan *tapApi.OutputChannelItem)
	filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)
	go FilterItems(outputItemsChannel, filteredOutputItemsChannel)
	go api.StartReadingEntries(filteredOutputItemsChannel, nil, ExtensionsMap)

	return outputItemsChannel
}

func FilterItems(inChannel <-chan *tapApi.OutputChannelItem, outChannel chan *tapApi.OutputChannelItem) {
	for message := range inChannel {
		if message.ConnectionInfo.IsOutgoing && api.CheckIsServiceIP(message.ConnectionInfo.ServerIP) {
			continue
		}

		outChannel <- message
	}
}
