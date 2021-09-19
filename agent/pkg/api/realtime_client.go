package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"mizuserver/pkg/models"
	"net"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func Connect(host string, port string) (conn net.Conn) {
	dest := host + ":" + port

	fmt.Printf("Connecting to %s...\n", dest)

	conn, err := net.Dial("tcp", dest)

	if err != nil {
		if _, t := err.(*net.OpError); t {
			fmt.Println("Some problem connecting.")
		} else {
			fmt.Println("Unknown error: " + err.Error())
		}
		os.Exit(1)
	}

	return
}

func SetModeInsert(conn net.Conn) {
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	conn.Write([]byte("/insert\n"))
}

func Insert(entry interface{}, conn net.Conn) {
	var data []byte
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	data, _ = json.Marshal(entry)
	conn.Write(data)

	conn.Write([]byte("\n"))
}

func Single(index uint) (entry map[string]interface{}) {
	conn := Connect("localhost", "8000")

	var wg sync.WaitGroup
	go readConnectionSingle(&wg, conn, &entry)
	wg.Add(1)

	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	conn.Write([]byte("/single\n"))

	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	conn.Write([]byte(fmt.Sprintf("%d\n", index)))

	wg.Wait()
	return
}

func readConnectionSingle(wg *sync.WaitGroup, conn net.Conn, entry *map[string]interface{}) {
	defer wg.Done()
	for {
		scanner := bufio.NewScanner(conn)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 209715200)

		for {
			ok := scanner.Scan()
			text := scanner.Text()

			command := handleCommands(text)
			if !command {
				// fmt.Printf("\b\b** %s\n> ", text)

				if text == "" {
					err := scanner.Err()
					fmt.Printf("err: %v\n", err)
					return
				}

				if err := json.Unmarshal([]byte(text), entry); err != nil {
					panic(err)
				}
				return
			}

			if !ok {
				fmt.Println("Reached EOF on server connection.")
				break
			}
		}
	}
}

func Query(query string, conn net.Conn, ws *websocket.Conn) {
	var wg sync.WaitGroup
	go readConnection(&wg, conn, ws)
	wg.Add(1)

	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	conn.Write([]byte("/query\n"))

	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	conn.Write([]byte(fmt.Sprintf("%s\n", query)))

	wg.Wait()
}

func readConnection(wg *sync.WaitGroup, conn net.Conn, ws *websocket.Conn) {
	defer wg.Done()
	for {
		scanner := bufio.NewScanner(conn)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 209715200)

		for {
			ok := scanner.Scan()
			text := scanner.Text()

			command := handleCommands(text)
			if !command {
				// fmt.Printf("\b\b** %s\n> ", text)

				if text == "" {
					err := scanner.Err()
					fmt.Printf("err: %v\n", err)
					return
				}

				var data map[string]interface{}
				if err := json.Unmarshal([]byte(text), &data); err != nil {
					panic(err)
				}

				summary := data["summary"].(map[string]interface{})
				summary["id"] = uint(data["id"].(float64))

				baseEntryBytes, _ := models.CreateBaseEntryWebSocketMessage(summary)
				ws.WriteMessage(1, baseEntryBytes)
			}

			if !ok {
				fmt.Println("Reached EOF on server connection.")
				break
			}
		}
	}
}

func handleCommands(text string) bool {
	r, err := regexp.Compile("^%.*%$")
	if err == nil {
		if r.MatchString(text) {

			switch {
			case text == "%quit%":
				fmt.Println("\b\bServer is leaving. Hanging up.")
				os.Exit(0)
			}

			return true
		}
	}

	return false
}
