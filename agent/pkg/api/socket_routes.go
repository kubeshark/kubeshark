package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/up9inc/mizu/agent/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

type EventHandlers interface {
	WebSocketConnect(socketId int, isTapper bool)
	WebSocketDisconnect(socketId int, isTapper bool)
	WebSocketMessage(socketId int, message []byte)
}

type SocketConnection struct {
	connection    *websocket.Conn
	lock          *sync.Mutex
	eventHandlers EventHandlers
	isTapper      bool
}

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	websocketIdsLock            = sync.Mutex{}
	connectedWebsockets         map[int]*SocketConnection
	connectedWebsocketIdCounter = 0
	SocketGetBrowserHandler     gin.HandlerFunc
	SocketGetTapperHandler      gin.HandlerFunc
)

func init() {
	websocketUpgrader.CheckOrigin = func(r *http.Request) bool { return true } // like cors for web socket
	connectedWebsockets = make(map[int]*SocketConnection)
}

func WebSocketRoutes(app *gin.Engine, eventHandlers EventHandlers, startTime int64) {
	SocketGetBrowserHandler = func(c *gin.Context) {
		websocketHandler(c.Writer, c.Request, eventHandlers, false, startTime)
	}

	SocketGetTapperHandler = func(c *gin.Context) {
		websocketHandler(c.Writer, c.Request, eventHandlers, true, startTime)
	}

	app.GET("/ws", func(c *gin.Context) {
		SocketGetBrowserHandler(c)
	})

	app.GET("/wsTapper", func(c *gin.Context) { // TODO: add m2m authentication to this route
		SocketGetTapperHandler(c)
	})
}

func websocketHandler(w http.ResponseWriter, r *http.Request, eventHandlers EventHandlers, isTapper bool, startTime int64) {
	ws, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Errorf("Failed to set websocket upgrade: %v", err)
		return
	}

	websocketIdsLock.Lock()

	connectedWebsocketIdCounter++
	socketId := connectedWebsocketIdCounter
	connectedWebsockets[socketId] = &SocketConnection{connection: ws, lock: &sync.Mutex{}, eventHandlers: eventHandlers, isTapper: isTapper}

	websocketIdsLock.Unlock()

	var connection *basenine.Connection
	var isQuerySet bool

	// `!isTapper` means it's a connection from the web UI
	if !isTapper {
		connection, err = basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
		if err != nil {
			panic(err)
		}
	}

	data := make(chan []byte)
	meta := make(chan []byte)

	defer func() {
		socketCleanup(socketId, connectedWebsockets[socketId])
		data <- []byte(basenine.CloseChannel)
		meta <- []byte(basenine.CloseChannel)
		connection.Close()
	}()

	eventHandlers.WebSocketConnect(socketId, isTapper)

	startTimeBytes, _ := models.CreateWebsocketStartTimeMessage(startTime)

	if err = SendToSocket(socketId, startTimeBytes); err != nil {
		logger.Log.Error(err)
	}

	for {
		// params[0]: query
		// params[1]: enableFullEntries (0: disable, 1: enable)
		params := make([][]byte, 2)
		breakWholeLoop := false
		for i, _ := range params {
			_, params[i], err = ws.ReadMessage()
			if err != nil {
				if _, ok := err.(*websocket.CloseError); ok {
					logger.Log.Debugf("Received websocket close message, socket id: %d", socketId)
				} else {
					logger.Log.Errorf("Error reading message, socket id: %d, error: %v", socketId, err)
				}

				breakWholeLoop = true
				break
			}
		}

		if breakWholeLoop {
			break
		}

		enableFullEntries := false
		if len(params[1]) > 0 && params[1][0] != 48 {
			enableFullEntries = true
		}

		if !isTapper && !isQuerySet {
			query := string(params[0])
			err = basenine.Validate(shared.BasenineHost, shared.BaseninePort, query)
			if err != nil {
				toastBytes, _ := models.CreateWebsocketToastMessage(&models.ToastMessage{
					Type:      "error",
					AutoClose: 5000,
					Text:      fmt.Sprintf("Syntax error: %s", err.Error()),
				})
				if err := SendToSocket(socketId, toastBytes); err != nil {
					logger.Log.Error(err)
				}
				break
			}

			isQuerySet = true

			handleDataChannel := func(c *basenine.Connection, data chan []byte) {
				for {
					bytes := <-data

					if string(bytes) == basenine.CloseChannel {
						return
					}

					var entry *tapApi.Entry
					err = json.Unmarshal(bytes, &entry)

					var message []byte
					if enableFullEntries {
						message, _ = models.CreateFullEntryWebSocketMessage(entry)
					} else {
						base := tapApi.Summarize(entry)
						message, _ = models.CreateBaseEntryWebSocketMessage(base)
					}

					if err := SendToSocket(socketId, message); err != nil {
						logger.Log.Error(err)
					}
				}
			}

			handleMetaChannel := func(c *basenine.Connection, meta chan []byte) {
				for {
					bytes := <-meta

					if string(bytes) == basenine.CloseChannel {
						return
					}

					var metadata *basenine.Metadata
					err = json.Unmarshal(bytes, &metadata)
					if err != nil {
						logger.Log.Debugf("Error recieving metadata: %v", err.Error())
					}

					metadataBytes, _ := models.CreateWebsocketQueryMetadataMessage(metadata)
					if err := SendToSocket(socketId, metadataBytes); err != nil {
						logger.Log.Error(err)
					}
				}
			}

			go handleDataChannel(connection, data)
			go handleMetaChannel(connection, meta)

			connection.Query(query, data, meta)
		} else {
			eventHandlers.WebSocketMessage(socketId, params[0])
		}
	}
}

func socketCleanup(socketId int, socketConnection *SocketConnection) {
	err := socketConnection.connection.Close()
	if err != nil {
		logger.Log.Errorf("Error closing socket connection for socket id %d: %v", socketId, err)
	}

	websocketIdsLock.Lock()
	connectedWebsockets[socketId] = nil
	websocketIdsLock.Unlock()

	socketConnection.eventHandlers.WebSocketDisconnect(socketId, socketConnection.isTapper)
}

func SendToSocket(socketId int, message []byte) error {
	socketObj := connectedWebsockets[socketId]
	if socketObj == nil {
		return fmt.Errorf("Socket %v is disconnected", socketId)
	}

	var sent = false
	time.AfterFunc(time.Second*5, func() {
		if !sent {
			logger.Log.Error("Socket timed out")
			socketCleanup(socketId, socketObj)
		}
	})

	socketObj.lock.Lock() // gorilla socket panics from concurrent writes to a single socket
	err := socketObj.connection.WriteMessage(1, message)
	socketObj.lock.Unlock()
	sent = true

	if err != nil {
		return fmt.Errorf("Failed to write message to socket %v, err: %w", socketId, err)
	}
	return nil
}
