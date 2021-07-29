package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"mizuserver/pkg/api"
	"mizuserver/pkg/models"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/sensitiveDataFiltering"
	"mizuserver/pkg/utils"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
)

var tapperMode = flag.Bool("tap", false, "Run in tapper mode without API")
var demoMode = flag.Bool("demo", false, "Run in Demo mode with API")
var apiServerMode = flag.Bool("api-server", false, "Run in API server mode with API")
var standaloneMode = flag.Bool("standalone", false, "Run in standalone tapper and API mode")
var apiServerAddress = flag.String("api-server-address", "", "Address of mizu API server")
var namespace = flag.String("namespace", "", "Resolve IPs if they belong to resources in this namespace (default is all)")

func main() {
	flag.Parse()
	hostMode := os.Getenv(shared.HostModeEnvVar) == "1"
	tapOpts := &tap.TapOpts{HostMode: hostMode}

	if !*tapperMode && !*apiServerMode && !*standaloneMode {
		panic("One of the flags --tap, --api or --standalone must be provided")
	}

	if *standaloneMode {
		api.StartResolving(*namespace)

		harOutputChannel, outboundLinkOutputChannel := tap.StartPassiveTapper(tapOpts)
		filteredHarChannel := make(chan *tap.OutputChannelItem)

		go filterHarItems(harOutputChannel, filteredHarChannel, getTrafficFilteringOptions())
		go api.StartReadingEntries(filteredHarChannel, nil, false)
		go api.StartReadingOutbound(outboundLinkOutputChannel)

		hostApi(nil)
	} else if *tapperMode {
		if *apiServerAddress == "" {
			panic("API server address must be provided with --api-server-address when using --tap")
		}

		tapTargets := getTapTargets()
		if tapTargets != nil {
			tap.SetFilterAuthorities(tapTargets)
			rlog.Infof("Filtering for the following authorities: %v", tap.GetFilterIPs())
		}

		harOutputChannel, outboundLinkOutputChannel := tap.StartPassiveTapper(tapOpts)

		socketConnection, err := shared.ConnectToSocketServer(*apiServerAddress, shared.DEFAULT_SOCKET_RETRIES, shared.DEFAULT_SOCKET_RETRY_SLEEP_TIME, false)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to socket server at %s %v", *apiServerAddress, err))
		}

		go pipeChannelToSocket(socketConnection, harOutputChannel)
		go api.StartReadingOutbound(outboundLinkOutputChannel)
	} else if *apiServerMode {
		api.StartResolving(*namespace)

		socketHarOutChannel := make(chan *tap.OutputChannelItem, 1000)
		filteredHarChannel := make(chan *tap.OutputChannelItem)

		go filterHarItems(socketHarOutChannel, filteredHarChannel, getTrafficFilteringOptions())
		if *demoMode {
			workdir := "./hars"
			go api.StartReadingEntries(filteredHarChannel, &workdir, true)
		} else {
			go api.StartReadingEntries(filteredHarChannel, nil, false)
		}

		hostApi(socketHarOutChannel)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	rlog.Info("Exiting")
}

func hostApi(socketHarOutputChannel chan<- *tap.OutputChannelItem) {
	app := gin.Default()

	app.GET("/echo", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World 👋!")
	})

	eventHandlers := api.RoutesEventHandlers{
		SocketHarOutChannel: socketHarOutputChannel,
	}

	app.Use(static.ServeRoot("/", "./site"))
	app.Use(CORSMiddleware()) // This has to be called after the static middleware, does not work if its called before

	routes.WebSocketRoutes(app, &eventHandlers)
	routes.EntriesRoutes(app)
	routes.MetadataRoutes(app)
	routes.NotFoundRoute(app)

	utils.StartServer(app)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
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

		if !filterOptions.DisableRedaction {
			sensitiveDataFiltering.FilterSensitiveInfoFromHarRequest(message, filterOptions)
		}

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
