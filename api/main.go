package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/websocket"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/api"
	"mizuserver/pkg/middleware"
	"mizuserver/pkg/models"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/sensitiveDataFiltering"
	"mizuserver/pkg/tap"
	"mizuserver/pkg/utils"
	"os"
	"os/signal"
)

var shouldTap = flag.Bool("tap", false, "Run in tapper mode without API")
var aggregator = flag.Bool("aggregator", false, "Run in aggregator mode with API")
var standalone = flag.Bool("standalone", false, "Run in standalone tapper and API mode")
var aggregatorAddress = flag.String("aggregator-address", "", "Address of mizu collector for tapping")


func main() {
	flag.Parse()

	if !*shouldTap && !*aggregator && !*standalone{
		panic("One of the flags --tap, --api or --standalone must be provided")
	}

	if *standalone {
		harOutputChannel := tap.StartPassiveTapper()
		filteredHarChannel := make(chan *tap.OutputChannelItem)
		go filterHarHeaders(harOutputChannel, filteredHarChannel, getFilteringOptions())
		go api.StartReadingEntries(filteredHarChannel, nil)
		hostApi(nil)
	} else if *shouldTap {
		if *aggregatorAddress == "" {
			panic("Aggregator address must be provided with --aggregator-address when using --tap")
		}

		tapTargets := getTapTargets()
		if tapTargets != nil {
			tap.HostAppAddresses = tapTargets
			fmt.Println("Filtering for the following addresses:", tap.HostAppAddresses)
		}

		harOutputChannel := tap.StartPassiveTapper()
		socketConnection, err := shared.ConnectToSocketServer(*aggregatorAddress, shared.DEFAULT_SOCKET_RETRIES, shared.DEFAULT_SOCKET_RETRY_SLEEP_TIME, false)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to socket server at %s %v", *aggregatorAddress, err))
		}
		go pipeChannelToSocket(socketConnection, harOutputChannel)
	} else if *aggregator {
		socketHarOutChannel := make(chan *tap.OutputChannelItem, 1000)
		filteredHarChannel := make(chan *tap.OutputChannelItem)
		go api.StartReadingEntries(filteredHarChannel, nil)
		go filterHarHeaders(socketHarOutChannel, filteredHarChannel, getFilteringOptions())
		hostApi(socketHarOutChannel)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Exiting")
}

func hostApi(socketHarOutputChannel chan<- *tap.OutputChannelItem) {
	app := fiber.New()


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

func getFilteringOptions() *shared.FilteringOptions {
	filteringOptionsJson := os.Getenv(shared.MizuFilteringOptionsEnvVar)
	if filteringOptionsJson == "" {
		return nil
	}
	var filteringOptions shared.FilteringOptions
	err := json.Unmarshal([]byte(filteringOptionsJson), &filteringOptions)
	if err != nil {
		panic(fmt.Sprintf("env var %s's value of %s is invalid! json must match the shared.FilteringOptions struct %v", shared.MizuFilteringOptionsEnvVar, filteringOptionsJson, err))
	}

	return &filteringOptions
}

func filterHarHeaders(inChannel <- chan *tap.OutputChannelItem, outChannel chan *tap.OutputChannelItem, filterOptions *shared.FilteringOptions) {
	for message := range inChannel {
		sensitiveDataFiltering.FilterSensitiveInfoFromHarRequest(message, filterOptions)
		outChannel <- message
	}
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
			fmt.Printf("error converting message to json %s, (%v,%+v)\n", err, err, err)
			continue
		}

		err = connection.WriteMessage(websocket.TextMessage, marshaledData)
		if err != nil {
			fmt.Printf("error sending message through socket server %s, (%v,%+v)\n", err, err, err)
			continue
		}
	}
}
