package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"mizuserver/pkg/api"
	"mizuserver/pkg/config"
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"
	"mizuserver/pkg/models"
	"mizuserver/pkg/oas"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/servicemap"
	"mizuserver/pkg/up9"
	"mizuserver/pkg/utils"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"plugin"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/antelman107/net-wait-go/wait"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap"
	tapApi "github.com/up9inc/mizu/tap/api"
)

var tapperMode = flag.Bool("tap", false, "Run in tapper mode without API")
var apiServerMode = flag.Bool("api-server", false, "Run in API server mode with API")
var standaloneMode = flag.Bool("standalone", false, "Run in standalone tapper and API mode")
var apiServerAddress = flag.String("api-server-address", "", "Address of mizu API server")
var namespace = flag.String("namespace", "", "Resolve IPs if they belong to resources in this namespace (default is all)")
var harsReaderMode = flag.Bool("hars-read", false, "Run in hars-read mode")
var harsDir = flag.String("hars-dir", "", "Directory to read hars from")

var extensions []*tapApi.Extension             // global
var extensionsMap map[string]*tapApi.Extension // global

var startTime int64

const (
	socketConnectionRetries    = 30
	socketConnectionRetryDelay = time.Second * 2
	socketHandshakeTimeout     = time.Second * 2
	uiIndexPath                = "./site/index.html"
)

func main() {
	logLevel := determineLogLevel()
	logger.InitLoggerStderrOnly(logLevel)
	flag.Parse()
	if err := config.LoadConfig(); err != nil {
		logger.Log.Fatalf("Error loading config file %v", err)
	}
	loadExtensions()

	if !*tapperMode && !*apiServerMode && !*standaloneMode && !*harsReaderMode {
		panic("One of the flags --tap, --api or --standalone or --hars-read must be provided")
	}

	if *standaloneMode {
		api.StartResolving(*namespace)

		outputItemsChannel := make(chan *tapApi.OutputChannelItem)
		filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)

		filteringOptions := getTrafficFilteringOptions()
		hostMode := os.Getenv(shared.HostModeEnvVar) == "1"
		tapOpts := &tap.TapOpts{HostMode: hostMode}
		tap.StartPassiveTapper(tapOpts, outputItemsChannel, extensions, filteringOptions)

		go filterItems(outputItemsChannel, filteredOutputItemsChannel)
		go api.StartReadingEntries(filteredOutputItemsChannel, nil, extensionsMap)

		hostApi(nil)
	} else if *tapperMode {
		logger.Log.Infof("Starting tapper, websocket address: %s", *apiServerAddress)
		if *apiServerAddress == "" {
			panic("API server address must be provided with --api-server-address when using --tap")
		}

		hostMode := os.Getenv(shared.HostModeEnvVar) == "1"
		tapOpts := &tap.TapOpts{HostMode: hostMode}
		tapTargets := getTapTargets()
		if tapTargets != nil {
			tapOpts.FilterAuthorities = tapTargets
			logger.Log.Infof("Filtering for the following authorities: %v", tapOpts.FilterAuthorities)
		}

		filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)

		filteringOptions := getTrafficFilteringOptions()
		tap.StartPassiveTapper(tapOpts, filteredOutputItemsChannel, extensions, filteringOptions)
		socketConnection, err := dialSocketWithRetry(*apiServerAddress, socketConnectionRetries, socketConnectionRetryDelay)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to socket server at %s %v", *apiServerAddress, err))
		}
		logger.Log.Infof("Connected successfully to websocket %s", *apiServerAddress)

		go pipeTapChannelToSocket(socketConnection, filteredOutputItemsChannel)
	} else if *apiServerMode {
		configureBasenineServer(shared.BasenineHost, shared.BaseninePort)
		startTime = time.Now().UnixNano() / int64(time.Millisecond)
		api.StartResolving(*namespace)

		outputItemsChannel := make(chan *tapApi.OutputChannelItem)
		filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)
		enableExpFeatureIfNeeded()
		go filterItems(outputItemsChannel, filteredOutputItemsChannel)
		go api.StartReadingEntries(filteredOutputItemsChannel, nil, extensionsMap)

		syncEntriesConfig := getSyncEntriesConfig()
		if syncEntriesConfig != nil {
			if err := up9.SyncEntries(syncEntriesConfig); err != nil {
				logger.Log.Error("Error syncing entries, err: %v", err)
			}
		}

		hostApi(outputItemsChannel)
	} else if *harsReaderMode {
		outputItemsChannel := make(chan *tapApi.OutputChannelItem, 1000)
		filteredHarChannel := make(chan *tapApi.OutputChannelItem)

		go filterItems(outputItemsChannel, filteredHarChannel)
		go api.StartReadingEntries(filteredHarChannel, harsDir, extensionsMap)
		hostApi(nil)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	logger.Log.Info("Exiting")
}

func enableExpFeatureIfNeeded() {
	if config.Config.OAS {
		oas.GetOasGeneratorInstance().Start()
	}
	if config.Config.ServiceMap {
		servicemap.GetInstance().SetConfig(config.Config)
	}
}

func configureBasenineServer(host string, port string) {
	if !wait.New(
		wait.WithProto("tcp"),
		wait.WithWait(200*time.Millisecond),
		wait.WithBreak(50*time.Millisecond),
		wait.WithDeadline(5*time.Second),
		wait.WithDebug(true),
	).Do([]string{fmt.Sprintf("%s:%s", host, port)}) {
		logger.Log.Panicf("Basenine is not available!")
	}

	// Limit the database size to default 200MB
	err := basenine.Limit(host, port, config.Config.MaxDBSizeBytes)
	if err != nil {
		logger.Log.Panicf("Error while limiting database size: %v", err)
	}

	// Define the macros
	for _, extension := range extensions {
		macros := extension.Dissector.Macros()
		for macro, expanded := range macros {
			err = basenine.Macro(host, port, macro, expanded)
			if err != nil {
				logger.Log.Panicf("Error while adding a macro: %v", err)
			}
		}
	}
}

func loadExtensions() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	extensionsDir := path.Join(dir, "./extensions/")

	files, err := ioutil.ReadDir(extensionsDir)
	if err != nil {
		logger.Log.Fatal(err)
	}
	extensions = make([]*tapApi.Extension, len(files))
	extensionsMap = make(map[string]*tapApi.Extension)
	for i, file := range files {
		filename := file.Name()
		logger.Log.Infof("Loading extension: %s", filename)
		extension := &tapApi.Extension{
			Path: path.Join(extensionsDir, filename),
		}
		plug, _ := plugin.Open(extension.Path)
		extension.Plug = plug
		symDissector, err := plug.Lookup("Dissector")

		var dissector tapApi.Dissector
		var ok bool
		dissector, ok = symDissector.(tapApi.Dissector)
		if err != nil || !ok {
			panic(fmt.Sprintf("Failed to load the extension: %s", extension.Path))
		}
		dissector.Register(extension)
		extension.Dissector = dissector
		extensions[i] = extension
		extensionsMap[extension.Protocol.Name] = extension
	}

	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Protocol.Priority < extensions[j].Protocol.Priority
	})

	for _, extension := range extensions {
		logger.Log.Infof("Extension Properties: %+v", extension)
	}

	controllers.InitExtensionsMap(extensionsMap)
}

func hostApi(socketHarOutputChannel chan<- *tapApi.OutputChannelItem) {
	app := gin.Default()

	app.GET("/echo", func(c *gin.Context) {
		c.String(http.StatusOK, "Here is Mizu agent")
	})

	eventHandlers := api.RoutesEventHandlers{
		SocketOutChannel: socketHarOutputChannel,
	}

	app.Use(DisableRootStaticCache())

	if err := setUIFlags(); err != nil {
		logger.Log.Errorf("Error setting ui mode, err: %v", err)
	}

	if config.Config.StandaloneMode {
		app.Use(static.ServeRoot("/", "./site-standalone"))
	} else {
		app.Use(static.ServeRoot("/", "./site"))
	}

	app.Use(middlewares.CORSMiddleware()) // This has to be called after the static middleware, does not work if its called before

	api.WebSocketRoutes(app, &eventHandlers, startTime)

	if config.Config.StandaloneMode {
		routes.ConfigRoutes(app)
		routes.UserRoutes(app)
		routes.InstallRoutes(app)
	}
	if config.Config.OAS {
		routes.OASRoutes(app)
	}
	if config.Config.ServiceMap {
		routes.ServiceMapRoutes(app)
	}

	routes.QueryRoutes(app)
	routes.EntriesRoutes(app)
	routes.MetadataRoutes(app)
	routes.StatusRoutes(app)
	routes.NotFoundRoute(app)
	utils.StartServer(app)
}

func DisableRootStaticCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/" {
			// Disable cache only for the main static route
			c.Writer.Header().Set("Cache-Control", "no-store")
		}

		c.Next()
	}
}

func setUIFlags() error {
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

func parseEnvVar(env string) map[string][]v1.Pod {
	var mapOfList map[string][]v1.Pod

	val, present := os.LookupEnv(env)

	if !present {
		return mapOfList
	}

	err := json.Unmarshal([]byte(val), &mapOfList)
	if err != nil {
		panic(fmt.Sprintf("env var %s's value of %v is invalid! must be map[string][]v1.Pod %v", env, mapOfList, err))
	}
	return mapOfList
}

func getTapTargets() []v1.Pod {
	nodeName := os.Getenv(shared.NodeNameEnvVar)
	tappedAddressesPerNodeDict := parseEnvVar(shared.TappedAddressesPerNodeDictEnvVar)
	return tappedAddressesPerNodeDict[nodeName]
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

func filterItems(inChannel <-chan *tapApi.OutputChannelItem, outChannel chan *tapApi.OutputChannelItem) {
	for message := range inChannel {
		if message.ConnectionInfo.IsOutgoing && api.CheckIsServiceIP(message.ConnectionInfo.ServerIP) {
			continue
		}

		outChannel <- message
	}
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

func getSyncEntriesConfig() *shared.SyncEntriesConfig {
	syncEntriesConfigJson := os.Getenv(shared.SyncEntriesConfigEnvVar)
	if syncEntriesConfigJson == "" {
		return nil
	}

	var syncEntriesConfig = &shared.SyncEntriesConfig{}
	err := json.Unmarshal([]byte(syncEntriesConfigJson), syncEntriesConfig)
	if err != nil {
		panic(fmt.Sprintf("env var %s's value of %s is invalid! json must match the shared.SyncEntriesConfig struct, err: %v", shared.SyncEntriesConfigEnvVar, syncEntriesConfigJson, err))
	}

	return syncEntriesConfig
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
				default:
					logger.Log.Warningf("Received socket message of type %s for which no handlers are defined", socketMessageBase.MessageType)
				}
			}
		}
	}
}
