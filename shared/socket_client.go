package shared

import (
	"github.com/gorilla/websocket"
	"time"
)

func ConnectToSocketServer(address string) (*websocket.Conn, error) {
	const maxTry = 5
	const sleepTime = time.Second * 2
	var err error
	var connection *websocket.Conn
	try := 0

	// Connection to server fails if client pod is up before server.
	// Retries solve this issue.
	for try < maxTry {
		connection, _, err = websocket.DefaultDialer.Dial(address, nil)
		if err != nil {
			try++
			// fmt.Printf("Failed connecting to websocket server: %s, (%v,%+v)\n", err, err, err)
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
