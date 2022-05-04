package api

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/dependency"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	tapApi "github.com/up9inc/mizu/tap/api"
)

type EntryStreamer interface {
	Get(ctx context.Context, socketId int, params *WebSocketParams) error
}

type BasenineEntryStreamer struct{}

type FetchLocker struct {
	sync.Mutex
}

func (e *BasenineEntryStreamer) Get(ctx context.Context, socketId int, params *WebSocketParams) error {
	var connection *basenine.Connection

	entryStreamerSocketConnector := dependency.GetInstance(dependency.EntryStreamerSocketConnector).(EntryStreamerSocketConnector)

	connection, err := basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
	if err != nil {
		logger.Log.Errorf("Failed to establish a connection to Basenine: %v", err)
		entryStreamerSocketConnector.CleanupSocket(socketId)
		return err
	}

	data := make(chan []byte)
	meta := make(chan []byte)

	query := params.Query
	err = basenine.Validate(shared.BasenineHost, shared.BaseninePort, query)
	if err != nil {
		entryStreamerSocketConnector.SendToastError(socketId, err)
	}

	fetchLocker := &FetchLocker{}
	if params.Fetch > 0 {
		fetchLocker.Lock()
	}

	handleDataChannel := func(c *basenine.Connection, data chan []byte) {
		for {
			bytes := <-data

			if string(bytes) == basenine.CloseChannel {
				return
			}

			if params.Fetch > 0 {
				fetchLocker.Lock()
			}

			var entry *tapApi.Entry
			err = json.Unmarshal(bytes, &entry)
			if err != nil {
				logger.Log.Debugf("Error unmarshalling entry: %v", err.Error())
				continue
			}

			entryStreamerSocketConnector.SendEntry(socketId, entry, params)
		}
	}

	handleMetaChannel := func(c *basenine.Connection, meta chan []byte) {
		for {
			bytes := <-meta

			if string(bytes) == basenine.CloseChannel {
				return
			}

			var metadata *basenine.Metadata
			err = json.Unmarshal(bytes, &metadata)
			if err != nil {
				logger.Log.Debugf("Error unmarshalling metadata: %v", err.Error())
				continue
			}

			if params.Fetch > 0 {
				var fetchData [][]byte
				var fetchMeta []byte
				fetchData, fetchMeta, err = basenine.Fetch(
					shared.BasenineHost,
					shared.BaseninePort,
					metadata.LeftOff,
					-1,
					query,
					params.Fetch,
					time.Duration(params.TimeoutMs)*time.Millisecond,
				)

				if err != nil {
					entryStreamerSocketConnector.SendMetadata(socketId, metadata)
					fetchLocker.Unlock()
					continue
				}

				fetchData = e.reverseByteSlice(fetchData)
				for _, row := range fetchData {
					var entry *tapApi.Entry
					err = json.Unmarshal(row, &entry)
					if err != nil {
						break
					}

					entryStreamerSocketConnector.SendEntry(socketId, entry, params)
				}

				var fetchMetadata *basenine.Metadata
				err = json.Unmarshal(fetchMeta, &fetchMetadata)
				if err == nil {
					entryStreamerSocketConnector.SendMetadata(socketId, fetchMetadata)
				}

				params.Fetch = 0
				fetchLocker.Unlock()
			}

			entryStreamerSocketConnector.SendMetadata(socketId, metadata)
		}
	}

	go handleDataChannel(connection, data)
	go handleMetaChannel(connection, meta)

	if err = connection.Query(query, data, meta); err != nil {
		logger.Log.Errorf("Query mode call failed: %v", err)
		entryStreamerSocketConnector.CleanupSocket(socketId)
		return err
	}

	go func() {
		<-ctx.Done()
		data <- []byte(basenine.CloseChannel)
		meta <- []byte(basenine.CloseChannel)
		connection.Close()
	}()

	return nil
}

// Reverses a []byte slice.
func (e *BasenineEntryStreamer) reverseByteSlice(arr [][]byte) (newArr [][]byte) {
	for i := len(arr) - 1; i >= 0; i-- {
		newArr = append(newArr, arr[i])
	}
	return newArr
}
