package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/up9inc/mizu/shared/debounce"
	"net/http"
	"sync"
	"time"
)

type EventHandlers interface {
	WebSocketConnect(socketId int, isTapper bool)
	WebSocketDisconnect(socketId int, isTapper bool)
	WebSocketMessage(socketId int, message []byte)
}

type SocketConnection struct {
	connection *websocket.Conn
	lock *sync.Mutex
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
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}

	websocketIdsLock.Lock()

	connectedWebsocketIdCounter++
	socketId := connectedWebsocketIdCounter
	connectedWebsockets[socketId] = &SocketConnection{connection: conn, lock: &sync.Mutex{}}

	websocketIdsLock.Unlock()

	defer func() {
		connectedWebsockets[socketId] = nil
		eventHandlers.WebSocketDisconnect(socketId, isTapper)
	}()

	eventHandlers.WebSocketConnect(socketId, isTapper)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Conn err: %v\n", err)
			break
		}
		eventHandlers.WebSocketMessage(socketId, msg)
	}
}

var db = debounce.NewDebouncer(time.Second * 5, func() {
	fmt.Println("Successfully sent to socket")
})

func SendToSocket(socketId int, message []byte) error {
	var sent = false
	socketObj := connectedWebsockets[socketId]
	time.AfterFunc(time.Second * 30, func() {
		if !sent {
			fmt.Printf("Write to socket id %d timed out after 30s, closing the socket\n", socketId)
			err := socketObj.connection.Close()
			if err != nil {
				fmt.Printf("Error closing connection for socket id %d:%v\n", socketId, err)
			}
		} else {
			db.SetOn()
		}
	})
	socketObj.lock.Lock() // gorilla socket panics from concurrent writes to a single socket
	writeRes := socketObj.connection.WriteMessage(1, message)
	socketObj.lock.Unlock()
	sent = true
	return writeRes
}
