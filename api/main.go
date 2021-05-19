package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/api"
	"mizuserver/pkg/middleware"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/tap"
	"mizuserver/pkg/utils"
	"os"
	"os/signal"
)

var shouldTap = flag.Bool("tap", false, "Run in tapper mode without API")
var aggregator = flag.Bool("aggregator", false, "Run in aggregator mode with API")
var standalone = flag.Bool("standalone", false, "Run in standalone tapper and API mode")
var aggregatorAddress = flag.String("aggregator-address", "", "Address of mizu collector for tapping")

const nodeNameEnvVar = "NODE_NAME"
const tappedAddressesPerNodeDictEnvVar = "TAPPED_ADDRESSES_PER_HOST"

func main() {
	flag.Parse()

	if !*shouldTap && !*aggregator && !*standalone{
		panic("One of the flags --tap, --api or --standalone must be provided")
	}

	if *standalone {
		harOutputChannel := tap.StartPassiveTapper()
		go api.StartReadingEntries(harOutputChannel, tap.HarOutputDir)
		hostApi()
	} else if *shouldTap {
		if *aggregatorAddress == "" {
			panic("Aggregator address must be provided with --aggregator-address when using --tap")
		}

		tapTargets := getTapTargets()
		if tapTargets != nil {
			tap.HostAppAddresses = tapTargets
		}

		harOutputChannel := tap.StartPassiveTapper()
		socketConnection, err := api.ConnectToSocketServer(*aggregatorAddress)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to socket server at %s %v", *aggregatorAddress, err))
		}
		go api.PipeChannelToSocket(socketConnection, harOutputChannel)
	} else if *aggregator {
		go api.StartReadingEntries(api.SocketHarOutChannel, nil)
		hostApi()
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Exiting")
}

func hostApi() {
	app := fiber.New()


	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.
	app.Static("/", "./site")

	//Simple route to know server is running
	app.Get("/echo", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	eventHandlers := RoutesEventHandlers{
		SocketHarOutChannel: make(chan *tap.OutputChannelItem, 1000),
	}
	routes.WebSocketRoutes(app, eventHandlers)
	routes.EntriesRoutes(app)
	routes.NotFoundRoute(app)

	utils.StartServer(app)
}


func getTapTargets() []string {
	nodeName := os.Getenv(nodeNameEnvVar)
	var tappedAddressesPerNodeDict map[string][]string
	err := json.Unmarshal([]byte(os.Getenv(tappedAddressesPerNodeDictEnvVar)), &tappedAddressesPerNodeDict)
	if err != nil {
		panic(fmt.Sprintf("env var value of %s is invalid! must be map[string][]string %v", tappedAddressesPerNodeDict, err))
	}
	return tappedAddressesPerNodeDict[nodeName]
}
