package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"mizuserver/pkg/models"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	"github.com/up9inc/mizu/shared/logger"
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

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var websocketIdsLock = sync.Mutex{}
var connectedWebsockets map[int]*SocketConnection
var connectedWebsocketIdCounter = 0

func init() {
	websocketUpgrader.CheckOrigin = func(r *http.Request) bool { return true } // like cors for web socket
	connectedWebsockets = make(map[int]*SocketConnection, 0)
}

func WebSocketRoutes(app *gin.Engine, eventHandlers EventHandlers, startTime int64) {
	app.GET("/ws", func(c *gin.Context) {
		websocketHandler(c.Writer, c.Request, eventHandlers, false, startTime)
	})
	app.GET("/wsTapper", func(c *gin.Context) {
		websocketHandler(c.Writer, c.Request, eventHandlers, true, startTime)
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
		data <- []byte(basenine.CloseChannel)
		meta <- []byte(basenine.CloseChannel)
		connection.Close()
		socketCleanup(socketId, connectedWebsockets[socketId])
	}()

	eventHandlers.WebSocketConnect(socketId, isTapper)

	startTimeBytes, _ := models.CreateWebsocketStartTimeMessage(startTime)
	BroadcastToBrowserClients(startTimeBytes)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			logger.Log.Errorf("Error reading message, socket id: %d, error: %v", socketId, err)
			break
		}

		if !isTapper && !isQuerySet {
			query := string(msg)
			err = basenine.Validate(shared.BasenineHost, shared.BaseninePort, query)
			if err != nil {
				toastBytes, _ := models.CreateWebsocketToastMessage(&models.ToastMessage{
					Type:      "error",
					AutoClose: 5000,
					Text:      fmt.Sprintf("Syntax error: %s", err.Error()),
				})
				BroadcastToBrowserClients(toastBytes)
				break
			}

			isQuerySet = true

			handleDataChannel := func(c *basenine.Connection, data chan []byte) {
				for {
					bytes := <-data

					if string(bytes) == basenine.CloseChannel {
						return
					}

					var dataMap map[string]interface{}
					err = json.Unmarshal(bytes, &dataMap)

					base := dataMap["base"].(map[string]interface{})
					base["id"] = uint(dataMap["id"].(float64))

					baseEntryBytes, _ := models.CreateBaseEntryWebSocketMessage(base)
					BroadcastToBrowserClients(baseEntryBytes)
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
						logger.Log.Debugf("Error recieving metadata: %v\n", err.Error())
					}

					metadataBytes, _ := models.CreateWebsocketQueryMetadataMessage(metadata)
					BroadcastToBrowserClients(metadataBytes)
				}
			}

			go handleDataChannel(connection, data)
			go handleMetaChannel(connection, meta)

			connection.Query(query, data, meta)
		} else {
			eventHandlers.WebSocketMessage(socketId, msg)
		}
	}
}

func socketCleanup(socketId int, socketConnection *SocketConnection) {
	err := socketConnection.connection.Close()
	if err != nil {
		logger.Log.Errorf("Error closing socket connection for socket id %d: %v\n", socketId, err)
	}

	websocketIdsLock.Lock()
	connectedWebsockets[socketId] = nil
	websocketIdsLock.Unlock()

	socketConnection.eventHandlers.WebSocketDisconnect(socketId, socketConnection.isTapper)
}

var db = debounce.NewDebouncer(time.Second*5, func() {
	logger.Log.Error("Successfully sent to socket")
})

func SendToSocket(socketId int, message []byte) error {
	socketObj := connectedWebsockets[socketId]
	if socketObj == nil {
		return errors.New("Socket is disconnected")
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
	return err
}
