package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/dependency"
	"github.com/up9inc/mizu/agent/pkg/entries"
	"github.com/up9inc/mizu/agent/pkg/middlewares"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/agent/pkg/oas"
	"github.com/up9inc/mizu/agent/pkg/routes"
	"github.com/up9inc/mizu/agent/pkg/servicemap"
	"github.com/up9inc/mizu/agent/pkg/utils"

	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/app"
	"github.com/up9inc/mizu/agent/pkg/config"

	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
	tapApi "github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/dbgctl"
)

var tapperMode = flag.Bool("tap", false, "Run in tapper mode without API")
var apiServerMode = flag.Bool("api-server", false, "Run in API server mode with API")
var standaloneMode = flag.Bool("standalone", false, "Run in standalone tapper and API mode")
var apiServerAddress = flag.String("api-server-address", "", "Address of mizu API server")
var namespace = flag.String("namespace", "", "Resolve IPs if they belong to resources in this namespace (default is all)")
var harsReaderMode = flag.Bool("hars-read", false, "Run in hars-read mode")
var harsDir = flag.String("hars-dir", "", "Directory to read hars from")
var profiler = flag.Bool("profiler", false, "Run pprof server")

const (
	socketConnectionRetries    = 30
	socketConnectionRetryDelay = time.Second * 2
	socketHandshakeTimeout     = time.Second * 2
)

func main() {
	initializeDependencies()
	logLevel := determineLogLevel()
	logger.InitLoggerStd(logLevel)
	flag.Parse()

	app.LoadExtensions()

	if !*tapperMode && !*apiServerMode && !*standaloneMode && !*harsReaderMode {
		panic("One of the flags --tap, --api-server, --standalone or --hars-read must be provided")
	}

	if *standaloneMode {
		runInStandaloneMode()
	} else if *tapperMode {
		runInTapperMode()
	} else if *apiServerMode {
		ginApp := runInApiServerMode(*namespace)

		if *profiler {
			pprof.Register(ginApp)
		}

		utils.StartServer(ginApp)

	} else if *harsReaderMode {
		runInHarReaderMode()
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	logger.Log.Info("Exiting")
}

func hostApi(socketHarOutputChannel chan<- *tapApi.OutputChannelItem) *gin.Engine {
	ginApp := gin.Default()

	ginApp.GET("/echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Here is Mizu agent")
	})

	eventHandlers := api.RoutesEventHandlers{
		SocketOutChannel: socketHarOutputChannel,
	}

	ginApp.Use(disableRootStaticCache())

	staticFolder := "./site"
	indexStaticFile := staticFolder + "/index.html"
	if err := setUIFlags(indexStaticFile); err != nil {
		logger.Log.Errorf("Error setting ui flags, err: %v", err)
	}

	ginApp.Use(static.ServeRoot("/", staticFolder))
	ginApp.NoRoute(func(c *gin.Context) {
		c.File(indexStaticFile)
	})

	ginApp.Use(middlewares.CORSMiddleware()) // This has to be called after the static middleware, does not work if it's called before

	api.WebSocketRoutes(ginApp, &eventHandlers)

	if config.Config.OAS {
		routes.OASRoutes(ginApp)
	}

	if config.Config.ServiceMap {
		routes.ServiceMapRoutes(ginApp)
	}

	routes.QueryRoutes(ginApp)
	routes.EntriesRoutes(ginApp)
	routes.MetadataRoutes(ginApp)
	routes.StatusRoutes(ginApp)
	routes.DbRoutes(ginApp)

	return ginApp
}

func runInApiServerMode(namespace string) *gin.Engine {
	if err := config.LoadConfig(); err != nil {
		logger.Log.Fatalf("Error loading config file %v", err)
	}
	app.ConfigureBasenineServer(shared.BasenineHost, shared.BaseninePort, config.Config.MaxDBSizeBytes, config.Config.LogLevel, config.Config.InsertionFilter)
	api.StartResolving(namespace)

	enableExpFeatureIfNeeded()

	return hostApi(app.GetEntryInputChannel())
}

func runInTapperMode() {
	logger.Log.Infof("Starting tapper, websocket address: %s", *apiServerAddress)
	if *apiServerAddress == "" {
		panic("API server address must be provided with --api-server-address when using --tap")
	}

	hostMode := os.Getenv(shared.HostModeEnvVar) == "1"
	tapOpts := &tap.TapOpts{
		HostMode:     hostMode,
	}

	filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)

	filteringOptions := getTrafficFilteringOptions()
	tap.StartPassiveTapper(tapOpts, filteredOutputItemsChannel, app.Extensions, filteringOptions)
	socketConnection, err := dialSocketWithRetry(*apiServerAddress, socketConnectionRetries, socketConnectionRetryDelay)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to socket server at %s %v", *apiServerAddress, err))
	}
	logger.Log.Infof("Connected successfully to websocket %s", *apiServerAddress)

	go pipeTapChannelToSocket(socketConnection, filteredOutputItemsChannel)
}

func runInStandaloneMode() {
	api.StartResolving(*namespace)

	outputItemsChannel := make(chan *tapApi.OutputChannelItem)
	filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)

	filteringOptions := getTrafficFilteringOptions()
	hostMode := os.Getenv(shared.HostModeEnvVar) == "1"
	tapOpts := &tap.TapOpts{HostMode: hostMode}
	tap.StartPassiveTapper(tapOpts, outputItemsChannel, app.Extensions, filteringOptions)

	go app.FilterItems(outputItemsChannel, filteredOutputItemsChannel)
	go api.StartReadingEntries(filteredOutputItemsChannel, nil, app.ExtensionsMap)

	ginApp := hostApi(nil)
	utils.StartServer(ginApp)
}

func runInHarReaderMode() {
	outputItemsChannel := make(chan *tapApi.OutputChannelItem, 1000)
	filteredHarChannel := make(chan *tapApi.OutputChannelItem)

	go app.FilterItems(outputItemsChannel, filteredHarChannel)
	go api.StartReadingEntries(filteredHarChannel, harsDir, app.ExtensionsMap)
	ginApp := hostApi(nil)
	utils.StartServer(ginApp)
}

func enableExpFeatureIfNeeded() {
	if config.Config.OAS {
		oasGenerator := dependency.GetInstance(dependency.OasGeneratorDependency).(oas.OasGenerator)
		oasGenerator.Start()
	}
	if config.Config.ServiceMap {
		serviceMapGenerator := dependency.GetInstance(dependency.ServiceMapGeneratorDependency).(servicemap.ServiceMap)
		serviceMapGenerator.Enable()
	}
}

func disableRootStaticCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/" {
			// Disable cache only for the main static route
			c.Writer.Header().Set("Cache-Control", "no-store")
		}

		c.Next()
	}
}

func setUIFlags(uiIndexPath string) error {
	read, err := ioutil.ReadFile(uiIndexPath)
	if err != nil {
		return err
	}

	replacedContent := strings.Replace(string(read), "__IS_OAS_ENABLED__", strconv.FormatBool(config.Config.OAS), 1)
	replacedContent = strings.Replace(replacedContent, "__IS_SERVICE_MAP_ENABLED__", strconv.FormatBool(config.Config.ServiceMap), 1)

	err = ioutil.WriteFile(uiIndexPath, []byte(replacedContent), 0)
	if err != nil {
		return err
	}

	return nil
}

func getTrafficFilteringOptions() *tapApi.TrafficFilteringOptions {
	filteringOptionsJson := os.Getenv(shared.MizuFilteringOptionsEnvVar)
	if filteringOptionsJson == "" {
		return &tapApi.TrafficFilteringOptions{
			IgnoredUserAgents: []string{},
		}
	}
	var filteringOptions tapApi.TrafficFilteringOptions
	err := json.Unmarshal([]byte(filteringOptionsJson), &filteringOptions)
	if err != nil {
		panic(fmt.Sprintf("env var %s's value of %s is invalid! json must match the api.TrafficFilteringOptions struct %v", shared.MizuFilteringOptionsEnvVar, filteringOptionsJson, err))
	}

	return &filteringOptions
}

func pipeTapChannelToSocket(connection *websocket.Conn, messageDataChannel <-chan *tapApi.OutputChannelItem) {
	if connection == nil {
		panic("Websocket connection is nil")
	}

	if messageDataChannel == nil {
		panic("Channel of captured messages is nil")
	}

	for messageData := range messageDataChannel {
		marshaledData, err := models.CreateWebsocketTappedEntryMessage(messageData)
		if err != nil {
			logger.Log.Errorf("error converting message to json %v, err: %s, (%v,%+v)", messageData, err, err, err)
			continue
		}

		if dbgctl.MizuTapperDisableSending {
			continue
		}

		// NOTE: This is where the `*tapApi.OutputChannelItem` leaves the code
		// and goes into the intermediate WebSocket.
		err = connection.WriteMessage(websocket.TextMessage, marshaledData)
		if err != nil {
			logger.Log.Errorf("error sending message through socket server %v, err: %s, (%v,%+v)", messageData, err, err, err)
			if errors.Is(err, syscall.EPIPE) {
				logger.Log.Warning("detected socket disconnection, reestablishing socket connection")
				connection, err = dialSocketWithRetry(*apiServerAddress, socketConnectionRetries, socketConnectionRetryDelay)
				if err != nil {
					logger.Log.Fatalf("error reestablishing socket connection: %v", err)
				} else {
					logger.Log.Info("recovered connection successfully")
				}
			}
			continue
		}
	}
}

func determineLogLevel() (logLevel logging.Level) {
	logLevel, err := logging.LogLevel(os.Getenv(shared.LogLevelEnvVar))
	if err != nil {
		logLevel = logging.INFO
	}

	return
}

func dialSocketWithRetry(socketAddress string, retryAmount int, retryDelay time.Duration) (*websocket.Conn, error) {
	var lastErr error
	dialer := &websocket.Dialer{ // we use our own dialer instead of the default due to the default's 45 sec handshake timeout, we occasionally encounter hanging socket handshakes when tapper tries to connect to api too soon
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: socketHandshakeTimeout,
	}
	for i := 1; i < retryAmount; i++ {
		socketConnection, _, err := dialer.Dial(socketAddress, nil)
		if err != nil {
			lastErr = err
			if i < retryAmount {
				logger.Log.Infof("socket connection to %s failed: %v, retrying %d out of %d in %d seconds...", socketAddress, err, i, retryAmount, retryDelay/time.Second)
				time.Sleep(retryDelay)
			}
		} else {
			go handleIncomingMessageAsTapper(socketConnection)
			return socketConnection, nil
		}
	}
	return nil, lastErr
}

func handleIncomingMessageAsTapper(socketConnection *websocket.Conn) {
	for {
		if _, message, err := socketConnection.ReadMessage(); err != nil {
			logger.Log.Errorf("error reading message from socket connection, err: %s, (%v,%+v)", err, err, err)
			if errors.Is(err, syscall.EPIPE) {
				// socket has disconnected, we can safely stop this goroutine
				return
			}
		} else {
			var socketMessageBase shared.WebSocketMessageMetadata
			if err := json.Unmarshal(message, &socketMessageBase); err != nil {
				logger.Log.Errorf("Could not unmarshal websocket message %v", err)
			} else {
				switch socketMessageBase.MessageType {
				case shared.WebSocketMessageTypeTapConfig:
					var tapConfigMessage *shared.WebSocketTapConfigMessage
					if err := json.Unmarshal(message, &tapConfigMessage); err != nil {
						logger.Log.Errorf("received unknown message from socket connection: %s, err: %s, (%v,%+v)", string(message), err, err, err)
					} else {
						tap.UpdateTapTargets(tapConfigMessage.TapTargets)
					}
				case shared.WebSocketMessageTypeUpdateTappedPods:
					var tappedPodsMessage shared.WebSocketTappedPodsMessage
					if err := json.Unmarshal(message, &tappedPodsMessage); err != nil {
						logger.Log.Infof("Could not unmarshal message of message type %s %v", socketMessageBase.MessageType, err)
						return
					}
					nodeName := os.Getenv(shared.NodeNameEnvVar)
					tap.UpdateTapTargets(tappedPodsMessage.NodeToTappedPodMap[nodeName])
				default:
					logger.Log.Warningf("Received socket message of type %s for which no handlers are defined", socketMessageBase.MessageType)
				}
			}
		}
	}
}

func initializeDependencies() {
	dependency.RegisterGenerator(dependency.ServiceMapGeneratorDependency, func() interface{} { return servicemap.GetDefaultServiceMapInstance() })
	dependency.RegisterGenerator(dependency.OasGeneratorDependency, func() interface{} { return oas.GetDefaultOasGeneratorInstance() })
	dependency.RegisterGenerator(dependency.EntriesInserter, func() interface{} { return api.GetBasenineEntryInserterInstance() })
	dependency.RegisterGenerator(dependency.EntriesProvider, func() interface{} { return &entries.BasenineEntriesProvider{} })
	dependency.RegisterGenerator(dependency.EntriesSocketStreamer, func() interface{} { return &api.BasenineEntryStreamer{} })
	dependency.RegisterGenerator(dependency.EntryStreamerSocketConnector, func() interface{} { return &api.DefaultEntryStreamerSocketConnector{} })
}
