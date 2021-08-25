package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mizuserver/pkg/api"
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/models"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/utils"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"plugin"
	"sort"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
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

func main() {
	flag.Parse()
	loadExtensions()
	hostMode := os.Getenv(shared.HostModeEnvVar) == "1"
	tapOpts := &tap.TapOpts{HostMode: hostMode}

	if !*tapperMode && !*apiServerMode && !*standaloneMode && !*harsReaderMode {
		panic("One of the flags --tap, --api or --standalone or --hars-read must be provided")
	}

	if *standaloneMode {
		api.StartResolving(*namespace)

		filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)
		tap.StartPassiveTapper(tapOpts, filteredOutputItemsChannel, extensions)

		// go filterHarItems(harOutputChannel, filteredOutputItemsChannel, getTrafficFilteringOptions())
		go api.StartReadingEntries(filteredOutputItemsChannel, nil, extensionsMap)
		// go api.StartReadingOutbound(outboundLinkOutputChannel)

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

		// harOutputChannel, outboundLinkOutputChannel := tap.StartPassiveTapper(tapOpts)
		filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)
		tap.StartPassiveTapper(tapOpts, filteredOutputItemsChannel, extensions)
		socketConnection, err := shared.ConnectToSocketServer(*apiServerAddress, shared.DEFAULT_SOCKET_RETRIES, shared.DEFAULT_SOCKET_RETRY_SLEEP_TIME, false)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to socket server at %s %v", *apiServerAddress, err))
		}

		go pipeTapChannelToSocket(socketConnection, filteredOutputItemsChannel)
		// go pipeOutboundLinksChannelToSocket(socketConnection, outboundLinkOutputChannel)
	} else if *apiServerMode {
		api.StartResolving(*namespace)

		socketHarOutChannel := make(chan *tapApi.OutputChannelItem, 1000)
		// TODO: filtered does not work
		// filteredHarChannel := make(chan *tapApi.OutputChannelItem)

		// go filterHarItems(socketHarOutChannel, filteredHarChannel, getTrafficFilteringOptions())
		go api.StartReadingEntries(socketHarOutChannel, nil, extensionsMap)

		hostApi(socketHarOutChannel)
	} else if *harsReaderMode {
		socketHarOutChannel := make(chan *tapApi.OutputChannelItem, 1000)
		// filteredHarChannel := make(chan *tap.OutputChannelItem)

		// go filterHarItems(socketHarOutChannel, filteredHarChannel, getTrafficFilteringOptions())
		go api.StartReadingEntries(socketHarOutChannel, harsDir, extensionsMap)
		hostApi(nil)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	rlog.Info("Exiting")
}

func loadExtensions() {
	appPorts := parseEnvVar(shared.AppPortsEnvVar)

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	extensionsDir := path.Join(dir, "./extensions/")

	files, err := ioutil.ReadDir(extensionsDir)
	if err != nil {
		log.Fatal(err)
	}
	extensions = make([]*tapApi.Extension, len(files))
	extensionsMap = make(map[string]*tapApi.Extension)
	for i, file := range files {
		filename := file.Name()
		log.Printf("Loading extension: %s\n", filename)
		extension := &tapApi.Extension{
			Path: path.Join(extensionsDir, filename),
		}
		plug, _ := plugin.Open(extension.Path)
		extension.Plug = plug
		symDissector, _ := plug.Lookup("Dissector")

		var dissector tapApi.Dissector
		dissector, _ = symDissector.(tapApi.Dissector)
		dissector.Register(extension)
		extension.Dissector = dissector
		extensions[i] = extension
		if ports, ok := appPorts[extension.Protocol.Name]; ok {
			log.Printf("Overriding \"%s\" extension's ports to: %v", extension.Protocol.Name, ports)
			extension.Protocol.Ports = ports
		}
		extensionsMap[extension.Protocol.Name] = extension
	}

	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Protocol.Priority < extensions[j].Protocol.Priority
	})

	for _, extension := range extensions {
		log.Printf("Extension Properties: %+v\n", extension)
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

	app.Use(static.ServeRoot("/", "./site"))
	app.Use(CORSMiddleware()) // This has to be called after the static middleware, does not work if its called before

	api.WebSocketRoutes(app, &eventHandlers)
	routes.EntriesRoutes(app)
	routes.MetadataRoutes(app)
	routes.StatusRoutes(app)
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

func parseEnvVar(env string) map[string][]string {
	var mapOfList map[string][]string

	val, present := os.LookupEnv(env)

	if !present {
		return mapOfList
	}

	err := json.Unmarshal([]byte(val), &mapOfList)
	if err != nil {
		panic(fmt.Sprintf("env var %s's value of %s is invalid! must be map[string][]string %v", env, mapOfList, err))
	}
	return mapOfList
}

func getTapTargets() []string {
	nodeName := os.Getenv(shared.NodeNameEnvVar)
	tappedAddressesPerNodeDict := parseEnvVar(shared.TappedAddressesPerNodeDictEnvVar)
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

// var userAgentsToFilter = []string{"kube-probe", "prometheus"}

//func filterHarItems(inChannel <-chan *tap.OutputChannelItem, outChannel chan *tap.OutputChannelItem, filterOptions *shared.TrafficFilteringOptions) {
//	for message := range inChannel {
//		if message.ConnectionInfo.IsOutgoing && api.CheckIsServiceIP(message.ConnectionInfo.ServerIP) {
//			continue
//		}
//		// TODO: move this to tappers https://up9.atlassian.net/browse/TRA-3441
//		if filterOptions.HideHealthChecks && isHealthCheckByUserAgent(message) {
//			continue
//		}
//
//		if !filterOptions.DisableRedaction {
//			sensitiveDataFiltering.FilterSensitiveInfoFromHarRequest(message, filterOptions)
//		}
//
//		outChannel <- message
//	}
//}

//func isHealthCheckByUserAgent(message *tap.OutputChannelItem) bool {
//	// for _, header := range message.HarEntry.Request.Headers {
//	// 	if strings.ToLower(header.Name) == "user-agent" {
//	// 		for _, userAgent := range userAgentsToFilter {
//	// 			if strings.Contains(strings.ToLower(header.Value), userAgent) {
//	// 				return true
//	// 			}
//	// 		}
//	// 		return false
//	// 	}
//	// }
//	return false
//}

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
			rlog.Infof("error converting message to json %s, (%v,%+v)\n", err, err, err)
			continue
		}

		// NOTE: This is where the `*tapApi.OutputChannelItem` leaves the code
		// and goes into the intermediate WebSocket.
		err = connection.WriteMessage(websocket.TextMessage, marshaledData)
		if err != nil {
			rlog.Infof("error sending message through socket server %s, (%v,%+v)\n", err, err, err)
			continue
		}
	}
}

func pipeOutboundLinksChannelToSocket(connection *websocket.Conn, outboundLinkChannel <-chan *tap.OutboundLink) {
	for outboundLink := range outboundLinkChannel {
		if outboundLink.SuggestedProtocol == tap.TLSProtocol {
			marshaledData, err := models.CreateWebsocketOutboundLinkMessage(outboundLink)
			if err != nil {
				rlog.Infof("Error converting outbound link to json %s, (%v,%+v)", err, err, err)
				continue
			}

			err = connection.WriteMessage(websocket.TextMessage, marshaledData)
			if err != nil {
				rlog.Infof("error sending outbound link message through socket server %s, (%v,%+v)", err, err, err)
				continue
			}
		}
	}
}
