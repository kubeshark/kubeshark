package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/kubeshark/agent/pkg/dependency"
	"github.com/kubeshark/kubeshark/agent/pkg/entries"
	"github.com/kubeshark/kubeshark/agent/pkg/middlewares"
	"github.com/kubeshark/kubeshark/agent/pkg/oas"
	"github.com/kubeshark/kubeshark/agent/pkg/routes"
	"github.com/kubeshark/kubeshark/agent/pkg/servicemap"
	"github.com/kubeshark/kubeshark/agent/pkg/utils"

	"github.com/kubeshark/kubeshark/agent/pkg/api"
	"github.com/kubeshark/kubeshark/agent/pkg/app"
	"github.com/kubeshark/kubeshark/agent/pkg/config"

	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared"
	tapApi "github.com/kubeshark/worker/api"
	"github.com/op/go-logging"
)

var namespace = flag.String("namespace", "", "Resolve IPs if they belong to resources in this namespace (default is all)")
var port = flag.Int("port", 80, "Port number of the HTTP server")
var profiler = flag.Bool("profiler", false, "Run pprof server")

func main() {
	initializeDependencies()
	logLevel := determineLogLevel()
	logger.InitLoggerStd(logLevel)
	flag.Parse()

	app.LoadExtensions()

	ginApp := runInApiServerMode(*namespace)

	if *profiler {
		pprof.Register(ginApp)
	}

	utils.StartServer(ginApp, *port)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	logger.Log.Info("Exiting")
}

func hostApi(socketHarOutputChannel chan<- *tapApi.OutputChannelItem) *gin.Engine {
	ginApp := gin.Default()

	ginApp.GET("/echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Here is Kubeshark agent")
	})

	eventHandlers := api.RoutesEventHandlers{
		SocketOutChannel: socketHarOutputChannel,
	}

	ginApp.Use(middlewares.CORSMiddleware())

	api.WebSocketRoutes(ginApp, &eventHandlers)

	if config.Config.OAS.Enable {
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
	routes.ReplayRoutes(ginApp)

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

func enableExpFeatureIfNeeded() {
	if config.Config.OAS.Enable {
		oasGenerator := dependency.GetInstance(dependency.OasGeneratorDependency).(oas.OasGenerator)
		oasGenerator.Start()
	}
	if config.Config.ServiceMap {
		serviceMapGenerator := dependency.GetInstance(dependency.ServiceMapGeneratorDependency).(servicemap.ServiceMap)
		serviceMapGenerator.Enable()
	}
}

func determineLogLevel() (logLevel logging.Level) {
	logLevel, err := logging.LogLevel(os.Getenv(shared.LogLevelEnvVar))
	if err != nil {
		logLevel = logging.INFO
	}

	return
}

func initializeDependencies() {
	dependency.RegisterGenerator(dependency.ServiceMapGeneratorDependency, func() interface{} { return servicemap.GetDefaultServiceMapInstance() })
	dependency.RegisterGenerator(dependency.OasGeneratorDependency, func() interface{} { return oas.GetDefaultOasGeneratorInstance(config.Config.OAS.MaxExampleLen) })
	dependency.RegisterGenerator(dependency.EntriesInserter, func() interface{} { return api.GetBasenineEntryInserterInstance() })
	dependency.RegisterGenerator(dependency.EntriesProvider, func() interface{} { return &entries.BasenineEntriesProvider{} })
	dependency.RegisterGenerator(dependency.EntriesSocketStreamer, func() interface{} { return &api.BasenineEntryStreamer{} })
	dependency.RegisterGenerator(dependency.EntryStreamerSocketConnector, func() interface{} { return &api.DefaultEntryStreamerSocketConnector{} })
}
