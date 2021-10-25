package api

import (
	"encoding/json"
	"errors"
	"mizuserver/pkg/models"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/romana/rlog"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
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

func WebSocketRoutes(app *gin.Engine, eventHandlers EventHandlers) {
	app.GET("/ws", func(c *gin.Context) {
		websocketHandler(c.Writer, c.Request, eventHandlers, false)
	})
	app.GET("/wsTapper", func(c *gin.Context) {
		websocketHandler(c.Writer, c.Request, eventHandlers, true)
	})
}

func websocketHandler(w http.ResponseWriter, r *http.Request, eventHandlers EventHandlers, isTapper bool) {
	ws, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		rlog.Errorf("Failed to set websocket upgrade: %v", err)
		return
	}

	websocketIdsLock.Lock()

	connectedWebsocketIdCounter++
	socketId := connectedWebsocketIdCounter
	connectedWebsockets[socketId] = &SocketConnection{connection: ws, lock: &sync.Mutex{}, eventHandlers: eventHandlers, isTapper: isTapper}

	websocketIdsLock.Unlock()

	var c *basenine.Connection
	var isQuerySet bool
	if !isTapper {
		c, err = basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
		if err != nil {
			panic(err)
		}
	}

	defer func() {
		c.Close()
		socketCleanup(socketId, connectedWebsockets[socketId])
	}()

	eventHandlers.WebSocketConnect(socketId, isTapper)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			rlog.Errorf("Error reading message, socket id: %d, error: %v", socketId, err)
			break
		}
		if !isTapper && !isQuerySet {
			isQuerySet = true
			data := make(chan []byte)

			handleDataChannel := func(wg *sync.WaitGroup, c *basenine.Connection, data chan []byte) {
				defer wg.Done()
				for {
					bytes := <-data

					var d map[string]interface{}
					err = json.Unmarshal(bytes, &d)

					summary := d["summary"].(map[string]interface{})
					summary["id"] = uint(d["id"].(float64))

					baseEntryBytes, _ := models.CreateBaseEntryWebSocketMessage(summary)
					ws.WriteMessage(1, baseEntryBytes)
				}
			}

			var wg sync.WaitGroup
			go handleDataChannel(&wg, c, data)
			wg.Add(1)

			c.Query(string(msg), data)

			wg.Wait()
		} else {
			eventHandlers.WebSocketMessage(socketId, msg)
		}
	}
}

func socketCleanup(socketId int, socketConnection *SocketConnection) {
	err := socketConnection.connection.Close()
	if err != nil {
		rlog.Errorf("Error closing socket connection for socket id %d: %v\n", socketId, err)
	}

	websocketIdsLock.Lock()
	connectedWebsockets[socketId] = nil
	websocketIdsLock.Unlock()

	socketConnection.eventHandlers.WebSocketDisconnect(socketId, socketConnection.isTapper)
}

var db = debounce.NewDebouncer(time.Second*5, func() {
	rlog.Error("Successfully sent to socket")
})

func SendToSocket(socketId int, message []byte) error {
	socketObj := connectedWebsockets[socketId]
	if socketObj == nil {
		return errors.New("Socket is disconnected")
	}

	var sent = false
	time.AfterFunc(time.Second*5, func() {
		if !sent {
			rlog.Error("Socket timed out")
			socketCleanup(socketId, socketObj)
		}
	})

	socketObj.lock.Lock() // gorilla socket panics from concurrent writes to a single socket
	err := socketObj.connection.WriteMessage(1, message)
	socketObj.lock.Unlock()

	sent = true
	return err
}
