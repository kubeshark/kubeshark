package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"mizuserver/pkg/tap"
	"time"
)

func ConnectToSocketServer(address string) (*websocket.Conn, error) {
	const maxTry = 3
	const sleepTime = time.Second * 10
	var err error
	var connection *websocket.Conn
	try := 0

	// Connection to server fails if client pod is up before server.
	// Retries solve this issue.
	for try < maxTry {
		connection, _, err = websocket.DefaultDialer.Dial(address, nil)
		if err != nil {
			try++
			fmt.Printf("Failed connecting to websocket server: %s, (%v,%+v)\n", err, err, err)
		} else {
			break
		}
		time.Sleep(sleepTime)
	}

	if err != nil {
		return nil, err
	}

	return connection, nil
}

func PipeChannelToSocket(connection *websocket.Conn, messageDataChannel <-chan *tap.OutputChannelItem) {
	if connection == nil {
		panic("Websocket connection is nil")
	}

	if messageDataChannel == nil {
		panic("Channel of captured messages is nil")
	}

	for messageData := range messageDataChannel {
		marshaledData, err := json.Marshal(messageData)
		if err != nil {
			fmt.Printf("error converting message to json %s, (%v,%+v)\n", err, err, err)
			continue
		}

		err = connection.WriteMessage(websocket.TextMessage, marshaledData)
		if err != nil {
			fmt.Printf("error sending message through socket server %s, (%v,%+v)\n", err, err, err)
			continue
		}
	}
}
