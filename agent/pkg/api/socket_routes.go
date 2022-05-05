package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/agent/pkg/utils"
	"github.com/up9inc/mizu/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

var extensionsMap map[string]*tapApi.Extension // global

func InitExtensionsMap(ref map[string]*tapApi.Extension) {
	extensionsMap = ref
}

type EventHandlers interface {
	WebSocketConnect(c *gin.Context, socketId int, isTapper bool)
	WebSocketDisconnect(socketId int, isTapper bool)
	WebSocketMessage(socketId int, isTapper bool, message []byte)
}

type SocketConnection struct {
	connection    *websocket.Conn
	lock          *sync.Mutex
	eventHandlers EventHandlers
	isTapper      bool
}

type WebSocketParams struct {
	LeftOff           string `json:"leftOff"`
	Query             string `json:"query"`
	EnableFullEntries bool   `json:"enableFullEntries"`
	Fetch             int    `json:"fetch"`
	TimeoutMs         int    `json:"timeoutMs"`
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

func WebSocketRoutes(app *gin.Engine, eventHandlers EventHandlers) {
	SocketGetBrowserHandler = func(c *gin.Context) {
		websocketHandler(c, eventHandlers, false)
	}

	SocketGetTapperHandler = func(c *gin.Context) {
		websocketHandler(c, eventHandlers, true)
	}

	app.GET("/ws", func(c *gin.Context) {
		SocketGetBrowserHandler(c)
	})

	app.GET("/wsTapper", func(c *gin.Context) { // TODO: add m2m authentication to this route
		SocketGetTapperHandler(c)
	})
}

func websocketHandler(c *gin.Context, eventHandlers EventHandlers, isTapper bool) {
	ws, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Errorf("failed to set websocket upgrade: %v", err)
		return
	}

	websocketIdsLock.Lock()

	connectedWebsocketIdCounter++
	socketId := connectedWebsocketIdCounter
	connectedWebsockets[socketId] = &SocketConnection{connection: ws, lock: &sync.Mutex{}, eventHandlers: eventHandlers, isTapper: isTapper}

	websocketIdsLock.Unlock()

	defer func() {
		socketCleanup(socketId, connectedWebsockets[socketId])
	}()

	eventHandlers.WebSocketConnect(c, socketId, isTapper)

	startTimeBytes, _ := models.CreateWebsocketStartTimeMessage(utils.StartTime)

	if err = SendToSocket(socketId, startTimeBytes); err != nil {
		logger.Log.Error(err)
	}

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				logger.Log.Debugf("received websocket close message, socket id: %d", socketId)
			} else {
				logger.Log.Errorf("error reading message, socket id: %d, error: %v", socketId, err)
			}

			break
		}

		eventHandlers.WebSocketMessage(socketId, isTapper, msg)
	}
}

func SendToSocket(socketId int, message []byte) error {
	socketObj := connectedWebsockets[socketId]
	if socketObj == nil {
		return fmt.Errorf("socket %v is disconnected", socketId)
	}

	socketObj.lock.Lock() // gorilla socket panics from concurrent writes to a single socket
	defer socketObj.lock.Unlock()

	if connectedWebsockets[socketId] == nil {
		return fmt.Errorf("socket %v is disconnected", socketId)
	}

	if err := socketObj.connection.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
		socketCleanup(socketId, socketObj)
		return fmt.Errorf("error setting timeout to socket %v, err: %v", socketId, err)
	}

	if err := socketObj.connection.WriteMessage(websocket.TextMessage, message); err != nil {
		socketCleanup(socketId, socketObj)
		return fmt.Errorf("failed to write message to socket %v, err: %v", socketId, err)
	}

	return nil
}

func socketCleanup(socketId int, socketConnection *SocketConnection) {
	err := socketConnection.connection.Close()
	if err != nil {
		logger.Log.Errorf("error closing socket connection for socket id %d: %v", socketId, err)
	}

	websocketIdsLock.Lock()
	connectedWebsockets[socketId] = nil
	websocketIdsLock.Unlock()

	socketConnection.eventHandlers.WebSocketDisconnect(socketId, socketConnection.isTapper)
}
