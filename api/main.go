package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gorilla/websocket"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
	"mizuserver/pkg/api"
	"mizuserver/pkg/middleware"
	"mizuserver/pkg/models"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/sensitiveDataFiltering"
	"mizuserver/pkg/utils"
	"os"
	"os/signal"
	"strings"
)

var shouldTap = flag.Bool("tap", false, "Run in tapper mode without API")
var aggregator = flag.Bool("aggregator", false, "Run in aggregator mode with API")
var standalone = flag.Bool("standalone", false, "Run in standalone tapper and API mode")
var aggregatorAddress = flag.String("aggregator-address", "", "Address of mizu collector for tapping")

func main() {
	flag.Parse()
	hostMode := os.Getenv(shared.HostModeEnvVar) == "1"
	tapOpts := &tap.TapOpts{HostMode: hostMode}

	if !*shouldTap && !*aggregator && !*standalone {
		panic("One of the flags --tap, --api or --standalone must be provided")
	}

	if *standalone {
		harOutputChannel, outboundLinkOutputChannel := tap.StartPassiveTapper(tapOpts)
		filteredHarChannel := make(chan *tap.OutputChannelItem)

		go filterHarItems(harOutputChannel, filteredHarChannel, getTrafficFilteringOptions())
		go api.StartReadingEntries(filteredHarChannel, nil)
		go api.StartReadingOutbound(outboundLinkOutputChannel)

		hostApi(nil)
	} else if *shouldTap {
		if *aggregatorAddress == "" {
			panic("Aggregator address must be provided with --aggregator-address when using --tap")
		}

		tapTargets := getTapTargets()
		if tapTargets != nil {
			tap.SetFilterAuthorities(tapTargets)
			rlog.Infof("Filtering for the following authorities: %v", tap.GetFilterIPs())
		}

		harOutputChannel, outboundLinkOutputChannel := tap.StartPassiveTapper(tapOpts)

		socketConnection, err := shared.ConnectToSocketServer(*aggregatorAddress, shared.DEFAULT_SOCKET_RETRIES, shared.DEFAULT_SOCKET_RETRY_SLEEP_TIME, false)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to socket server at %s %v", *aggregatorAddress, err))
		}

		go pipeChannelToSocket(socketConnection, harOutputChannel)
		go api.StartReadingOutbound(outboundLinkOutputChannel)
	} else if *aggregator {
		socketHarOutChannel := make(chan *tap.OutputChannelItem, 1000)
		filteredHarChannel := make(chan *tap.OutputChannelItem)

		go filterHarItems(socketHarOutChannel, filteredHarChannel, getTrafficFilteringOptions())
		go api.StartReadingEntries(filteredHarChannel, nil)

		hostApi(socketHarOutChannel)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	rlog.Info("Exiting")
}

func hostApi(socketHarOutputChannel chan<- *tap.OutputChannelItem) {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "*",
		AllowHeaders: "*",
	}))
	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.
	app.Static("/", "./site")

	//Simple route to know server is running
	app.Get("/echo", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})
	eventHandlers := api.RoutesEventHandlers{
		SocketHarOutChannel: socketHarOutputChannel,
	}
	routes.WebSocketRoutes(app, &eventHandlers)
	routes.EntriesRoutes(app)
	routes.MetadataRoutes(app)
	routes.NotFoundRoute(app)

	utils.StartServer(app)
}

func getTapTargets() []string {
	nodeName := os.Getenv(shared.NodeNameEnvVar)
	var tappedAddressesPerNodeDict map[string][]string
	err := json.Unmarshal([]byte(os.Getenv(shared.TappedAddressesPerNodeDictEnvVar)), &tappedAddressesPerNodeDict)
	if err != nil {
		panic(fmt.Sprintf("env var %s's value of %s is invalid! must be map[string][]string %v", shared.TappedAddressesPerNodeDictEnvVar, tappedAddressesPerNodeDict, err))
	}
	return tappedAddressesPerNodeDict[nodeName]
}

func getTrafficFilteringOptions() *shared.TrafficFilteringOptions {
	filteringOptionsJson := os.Getenv(shared.MizuFilteringOptionsEnvVar)
	if filteringOptionsJson == "" {
		return nil
	}
	var filteringOptions shared.TrafficFilteringOptions
	err := json.Unmarshal([]byte(filteringOptionsJson), &filteringOptions)
	if err != nil {
		panic(fmt.Sprintf("env var %s's value of %s is invalid! json must match the shared.TrafficFilteringOptions struct %v", shared.MizuFilteringOptionsEnvVar, filteringOptionsJson, err))
	}

	return &filteringOptions
}

var userAgentsToFilter = []string{"kube-probe", "prometheus"}

func filterHarItems(inChannel <-chan *tap.OutputChannelItem, outChannel chan *tap.OutputChannelItem, filterOptions *shared.TrafficFilteringOptions) {
	for message := range inChannel {
		if message.ConnectionInfo.IsOutgoing && api.CheckIsServiceIP(message.ConnectionInfo.ServerIP) {
			continue
		}
		// TODO: move this to tappers https://up9.atlassian.net/browse/TRA-3441
		if filterOptions.HideHealthChecks && isHealthCheckByUserAgent(message) {
			continue
		}

		sensitiveDataFiltering.FilterSensitiveInfoFromHarRequest(message, filterOptions)

		outChannel <- message
	}
}

func isHealthCheckByUserAgent(message *tap.OutputChannelItem) bool {
	for _, header := range message.HarEntry.Request.Headers {
		if strings.ToLower(header.Name) == "user-agent" {
			for _, userAgent := range userAgentsToFilter {
				if strings.Contains(strings.ToLower(header.Value), userAgent) {
					return true
				}
			}
			return false
		}
	}
	return false
}

func pipeChannelToSocket(connection *websocket.Conn, messageDataChannel <-chan *tap.OutputChannelItem) {
	if connection == nil {
		panic("Websocket connection is nil")
	}

	if messageDataChannel == nil {
		panic("Channel of captured messages is nil")
	}

	for messageData := range messageDataChannel {
		marshaledData, err := models.CreateWebsocketTappedEntryMessage(messageData)
		if err != nil {
			rlog.Infof("error converting message to json %s, (%v,%+v)\n", err, err, err)
			continue
		}

		err = connection.WriteMessage(websocket.TextMessage, marshaledData)
		if err != nil {
			rlog.Infof("error sending message through socket server %s, (%v,%+v)\n", err, err, err)
			continue
		}
	}
}
