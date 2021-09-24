package utils

import (
	"github.com/gorilla/websocket"
)

func ConnectToSocketServer(address string) (conn *websocket.Conn, err error) {
	conn, _, err = websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
