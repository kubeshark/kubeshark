package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/antelman107/net-wait-go/wait"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/config"
	"github.com/up9inc/mizu/agent/pkg/elastic"
	"github.com/up9inc/mizu/agent/pkg/middlewares"
	"github.com/up9inc/mizu/agent/pkg/oas"
	"github.com/up9inc/mizu/agent/pkg/routes"
	"github.com/up9inc/mizu/agent/pkg/servicemap"
	"github.com/up9inc/mizu/agent/pkg/up9"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

var (
	ConfigRoutes     *gin.RouterGroup
	UserRoutes       *gin.RouterGroup
	InstallRoutes    *gin.RouterGroup
	OASRoutes        *gin.RouterGroup
	ServiceMapRoutes *gin.RouterGroup
	QueryRoutes      *gin.RouterGroup
	EntriesRoutes    *gin.RouterGroup
	MetadataRoutes   *gin.RouterGroup
	StatusRoutes     *gin.RouterGroup

	startTime int64
)

func HostApi(socketHarOutputChannel chan<- *tapApi.OutputChannelItem) *gin.Engine {
	app := gin.Default()

	app.GET("/echo", func(c *gin.Context) {
		c.String(http.StatusOK, "Here is Mizu agent")
	})

	eventHandlers := api.RoutesEventHandlers{
		SocketOutChannel: socketHarOutputChannel,
	}

	app.Use(disableRootStaticCache())

	var staticFolder string
	if config.Config.StandaloneMode {
		staticFolder = "./site-standalone"
	} else {
		staticFolder = "./site"
	}

	indexStaticFile := staticFolder + "/index.html"
	if err := setUIFlags(indexStaticFile); err != nil {
		logger.Log.Errorf("Error setting ui flags, err: %v", err)
	}

	app.Use(static.ServeRoot("/", staticFolder))
	app.NoRoute(func(c *gin.Context) {
		c.File(indexStaticFile)
	})

	app.Use(middlewares.CORSMiddleware()) // This has to be called after the static middleware, does not work if its called before

	api.WebSocketRoutes(app, &eventHandlers, startTime)

	if config.Config.StandaloneMode {
		ConfigRoutes = routes.ConfigRoutes(app)
		UserRoutes = routes.UserRoutes(app)
		InstallRoutes = routes.InstallRoutes(app)
	}
	if config.Config.OAS {
		OASRoutes = routes.OASRoutes(app)
	}
	if config.Config.ServiceMap {
		ServiceMapRoutes = routes.ServiceMapRoutes(app)
	}

	QueryRoutes = routes.QueryRoutes(app)
	EntriesRoutes = routes.EntriesRoutes(app)
	MetadataRoutes = routes.MetadataRoutes(app)
	StatusRoutes = routes.StatusRoutes(app)

	return app
}

func RunInApiServerMode(namespace string) *gin.Engine {
	configureBasenineServer(shared.BasenineHost, shared.BaseninePort)
	startTime = time.Now().UnixNano() / int64(time.Millisecond)
	api.StartResolving(namespace)

	outputItemsChannel := make(chan *tapApi.OutputChannelItem)
	filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)
	enableExpFeatureIfNeeded()
	go FilterItems(outputItemsChannel, filteredOutputItemsChannel)
	go api.StartReadingEntries(filteredOutputItemsChannel, nil, ExtensionsMap)

	syncEntriesConfig := getSyncEntriesConfig()
	if syncEntriesConfig != nil {
		if err := up9.SyncEntries(syncEntriesConfig); err != nil {
			logger.Log.Error("Error syncing entries, err: %v", err)
		}
	}

	return HostApi(outputItemsChannel)
}

func configureBasenineServer(host string, port string) {
	if !wait.New(
		wait.WithProto("tcp"),
		wait.WithWait(200*time.Millisecond),
		wait.WithBreak(50*time.Millisecond),
		wait.WithDeadline(5*time.Second),
		wait.WithDebug(config.Config.LogLevel == logging.DEBUG),
	).Do([]string{fmt.Sprintf("%s:%s", host, port)}) {
		logger.Log.Panicf("Basenine is not available!")
	}

	// Limit the database size to default 200MB
	err := basenine.Limit(host, port, config.Config.MaxDBSizeBytes)
	if err != nil {
		logger.Log.Panicf("Error while limiting database size: %v", err)
	}

	// Define the macros
	for _, extension := range Extensions {
		macros := extension.Dissector.Macros()
		for macro, expanded := range macros {
			err = basenine.Macro(host, port, macro, expanded)
			if err != nil {
				logger.Log.Panicf("Error while adding a macro: %v", err)
			}
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

func FilterItems(inChannel <-chan *tapApi.OutputChannelItem, outChannel chan *tapApi.OutputChannelItem) {
	for message := range inChannel {
		if message.ConnectionInfo.IsOutgoing && api.CheckIsServiceIP(message.ConnectionInfo.ServerIP) {
			continue
		}

		outChannel <- message
	}
}

func enableExpFeatureIfNeeded() {
	if config.Config.OAS {
		oas.GetOasGeneratorInstance().Start()
	}
	if config.Config.ServiceMap {
		servicemap.GetInstance().SetConfig(config.Config)
	}
	elastic.GetInstance().Configure(config.Config.Elastic)
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
