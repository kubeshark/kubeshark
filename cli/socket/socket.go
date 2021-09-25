package socket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/up9inc/mizu/cli/logger"
)

const addr = "localhost:8898"

var upgrader = websocket.Upgrader{} // use default options

type SocketMessage struct {
	Type      string
	AutoClose uint
	Text      string
}

var socketMessageChannel chan *SocketMessage

func handle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Debugf("Error on WebSocket upgrade:", err)
		return
	}
	defer c.Close()
	for {
		msg := <-socketMessageChannel
		data, err := json.Marshal(msg)
		if err != nil {
			logger.Log.Debugf(err.Error())
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			logger.Log.Debugf("Error on WebSocket write:", err)
			break
		}
	}
}

func Listen() {
	socketMessageChannel = make(chan *SocketMessage)
	http.HandleFunc("/wsCli", handle)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func Send(_type string, autoClose uint, text string) {
	socketMessageChannel <- &SocketMessage{
		Type:      _type,
		AutoClose: autoClose,
		Text:      text,
	}
}
