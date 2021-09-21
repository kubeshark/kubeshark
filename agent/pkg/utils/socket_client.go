package utils

import (
	"github.com/gorilla/websocket"
	"github.com/romana/rlog"
	"time"
)

const (
	DEFAULT_SOCKET_RETRIES          = 3
	DEFAULT_SOCKET_RETRY_SLEEP_TIME = time.Second * 10
)

func ConnectToSocketServer(address string) (*websocket.Conn, error) {
	var err error
	var connection *websocket.Conn
	try := 0

	// Connection to server fails if client pod is up before server.
	// Retries solve this issue.
	for try < DEFAULT_SOCKET_RETRIES {
		rlog.Infof("Trying to connect to websocket: %s, attempt: %v/%v", address, try, DEFAULT_SOCKET_RETRIES)
		connection, _, err = websocket.DefaultDialer.Dial(address, nil)
		if err != nil {
			rlog.Warnf("Failed connecting to websocket: %s, attempt: %v/%v, err: %s, (%v,%+v)", address, try, DEFAULT_SOCKET_RETRIES, err, err, err)
			try++
		} else {
			break
		}
		time.Sleep(DEFAULT_SOCKET_RETRY_SLEEP_TIME)
	}

	if err != nil {
		return nil, err
	}

	return connection, nil
}
