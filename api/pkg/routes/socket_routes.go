package routes

import (
	"errors"
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
	connected bool
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
	connectedWebsockets[socketId] = &SocketConnection{connection: conn, lock: &sync.Mutex{}, connected: true}

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
	socketObj := connectedWebsockets[socketId]
	if socketObj == nil {
		return errors.New("Socket is disconnected")
	}

	var sent = false
	time.AfterFunc(time.Second * 5, func() {
		if !sent {
			fmt.Println("Socket timed out")
			connectedWebsockets[socketId] = nil
		}
	})

	socketObj.lock.Lock() // gorilla socket panics from concurrent writes to a single socket
	err := socketObj.connection.WriteMessage(1, message)
	socketObj.lock.Unlock()

	if err != nil {
		fmt.Printf("writeRes = %v\n", err)
	}

	sent = true
	return err
}
