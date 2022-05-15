package api

import (
	"encoding/json"
	"fmt"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap/api"
	"sync"
	"time"
)

type EntryInserter interface {
	Insert(entry *api.Entry) error
}

type BasenineEntryInserter struct {
	connection *basenine.Connection
}

var instance *BasenineEntryInserter
var once sync.Once

func GetBasenineEntryInserterInstance() *BasenineEntryInserter {
	once.Do(func() {
		instance = &BasenineEntryInserter{}
	})

	return instance
}

func (e *BasenineEntryInserter) Insert(entry *api.Entry) error {
	if e.connection == nil {
		e.connection = initializeConnection()
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("error marshling entry, err: %v", err)
	}

	if err := e.connection.SendText(string(data)); err != nil {
		e.connection.Close()
		e.connection = nil

		return fmt.Errorf("error sending text to database, err: %v", err)
	}

	return nil
}

func initializeConnection() *basenine.Connection{
	for {
		connection, err := basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
		if err != nil {
			logger.Log.Errorf("Can't establish a new connection to Basenine server: %v", err)
			time.Sleep(shared.BasenineReconnectInterval * time.Second)
			continue
		}

		if err = connection.InsertMode(); err != nil {
			logger.Log.Errorf("Insert mode call failed: %v", err)
			connection.Close()
			time.Sleep(shared.BasenineReconnectInterval * time.Second)
			continue
		}

		return connection
	}
}
