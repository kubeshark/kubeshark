package api

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/romana/rlog"
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
		query := c.DefaultQuery("q", "")
		websocketHandlerUI(c.Writer, c.Request, eventHandlers, false, query)
	})
	app.GET("/wsTapper", func(c *gin.Context) {
		websocketHandler(c.Writer, c.Request, eventHandlers, true)
	})
}

func websocketHandlerUI(w http.ResponseWriter, r *http.Request, eventHandlers EventHandlers, isTapper bool, query string) {
	ws, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		rlog.Errorf("Failed to set websocket upgrade: %v", err)
		return
	}

	conn := Connect("localhost", "8000")
	go Query(query, conn, ws)

	websocketIdsLock.Lock()

	connectedWebsocketIdCounter++
	socketId := connectedWebsocketIdCounter
	connectedWebsockets[socketId] = &SocketConnection{connection: ws, lock: &sync.Mutex{}, eventHandlers: eventHandlers, isTapper: isTapper}

	websocketIdsLock.Unlock()

	defer func() {
		socketCleanup(socketId, connectedWebsockets[socketId])
	}()

	eventHandlers.WebSocketConnect(socketId, isTapper)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			rlog.Errorf("Error reading message, socket id: %d, error: %v", socketId, err)
			break
		}
		eventHandlers.WebSocketMessage(socketId, msg)
	}

	conn.Close()
}

func websocketHandler(w http.ResponseWriter, r *http.Request, eventHandlers EventHandlers, isTapper bool) {
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		rlog.Errorf("Failed to set websocket upgrade: %v", err)
		return
	}

	websocketIdsLock.Lock()

	connectedWebsocketIdCounter++
	socketId := connectedWebsocketIdCounter
	connectedWebsockets[socketId] = &SocketConnection{connection: conn, lock: &sync.Mutex{}, eventHandlers: eventHandlers, isTapper: isTapper}

	websocketIdsLock.Unlock()

	defer func() {
		socketCleanup(socketId, connectedWebsockets[socketId])
	}()

	eventHandlers.WebSocketConnect(socketId, isTapper)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			rlog.Errorf("Error reading message, socket id: %d, error: %v", socketId, err)
			break
		}
		eventHandlers.WebSocketMessage(socketId, msg)
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
