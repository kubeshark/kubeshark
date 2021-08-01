package shared

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	DEFAULT_SOCKET_RETRIES          = 3
	DEFAULT_SOCKET_RETRY_SLEEP_TIME = time.Second * 10
)

func ConnectToSocketServer(address string, retries int, retrySleepTime time.Duration, hideTimeoutErrors bool) (*websocket.Conn, error) {
	var err error
	var connection *websocket.Conn
	try := 0

	// Connection to server fails if client pod is up before server.
	// Retries solve this issue.
	for try < retries {
		connection, _, err = websocket.DefaultDialer.Dial(address, nil)
		if err != nil {
			try++
			if !hideTimeoutErrors {
				fmt.Printf("Failed connecting to websocket server: %s, (%v,%+v)\n", err, err, err)
			}
		} else {
			break
		}
		time.Sleep(retrySleepTime)
	}

	if err != nil {
		return nil, err
	}

	return connection, nil
}
