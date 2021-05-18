package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"mizuserver/pkg/tap"
)

func ConnectToSocketServer(address string) (*websocket.Conn, error) {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

func PipeChannelToSocket(connection *websocket.Conn, messageDataChannel chan *tap.OutputChannelItem) {
	if connection == nil {
		panic("Websocket connection is nil")
	}

	if messageDataChannel == nil {
		panic("Channel of captured messages is nil")
	}

	for messageData := range messageDataChannel {
		marshaledData, err := json.Marshal(messageData)
		if err != nil {
			fmt.Printf("error converting message to json %v", err)
		}
		err = connection.WriteMessage(1, marshaledData)
		if err != nil {
			fmt.Printf("error sending message through socket server %v", err)
		}
	}
}
