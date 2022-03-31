package api

import (
	"context"
	"encoding/json"
	"fmt"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/dependency"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

type EntryStreamer interface {
	Get(socketId int, params *WebSocketParams) (context.CancelFunc, error)
}

type EntryStreamerSocketConnector interface {
	SendEntry(socketId int, entry *tapApi.Entry, params *WebSocketParams)
	SendMetadata(socketId int, metadata *basenine.Metadata)
	SendToastError(socketId int, err error)
	CleanupSocket(socketId int)
}

type DefaultEntryStreamerSocketConnector struct{}

func (e *DefaultEntryStreamerSocketConnector) SendEntry(socketId int, entry *tapApi.Entry, params *WebSocketParams) {
	var message []byte
	if params.EnableFullEntries {
		message, _ = models.CreateFullEntryWebSocketMessage(entry)
	} else {
		extension := extensionsMap[entry.Protocol.Name]
		base := extension.Dissector.Summarize(entry)
		message, _ = models.CreateBaseEntryWebSocketMessage(base)
	}

	if err := SendToSocket(socketId, message); err != nil {
		logger.Log.Error(err)
	}
}

func (e *DefaultEntryStreamerSocketConnector) SendMetadata(socketId int, metadata *basenine.Metadata) {
	metadataBytes, _ := models.CreateWebsocketQueryMetadataMessage(metadata)
	if err := SendToSocket(socketId, metadataBytes); err != nil {
		logger.Log.Error(err)
	}
}

func (e *DefaultEntryStreamerSocketConnector) SendToastError(socketId int, err error) {
	toastBytes, _ := models.CreateWebsocketToastMessage(&models.ToastMessage{
		Type:      "error",
		AutoClose: 5000,
		Text:      fmt.Sprintf("Syntax error: %s", err.Error()),
	})
	if err := SendToSocket(socketId, toastBytes); err != nil {
		logger.Log.Error(err)
	}
}

func (e *DefaultEntryStreamerSocketConnector) CleanupSocket(socketId int) {
	socketObj := connectedWebsockets[socketId]
	socketCleanup(socketId, socketObj)
}

type BasenineEntryStreamer struct{}

func (e *BasenineEntryStreamer) Get(socketId int, params *WebSocketParams) (context.CancelFunc, error) {
	var connection *basenine.Connection

	entryStreamerSocketConnector := dependency.GetInstance(dependency.EntryStreamerSocketConnector).(EntryStreamerSocketConnector)

	connection, err := basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
	if err != nil {
		logger.Log.Errorf("failed to establish a connection to Basenine: %v", err)
		entryStreamerSocketConnector.CleanupSocket(socketId)
		return nil, err
	}

	data := make(chan []byte)
	meta := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())

	query := params.Query
	err = basenine.Validate(shared.BasenineHost, shared.BaseninePort, query)
	if err != nil {
		entryStreamerSocketConnector.SendToastError(socketId, err)
	}

	handleDataChannel := func(c *basenine.Connection, data chan []byte) {
		for {
			bytes := <-data

			if string(bytes) == basenine.CloseChannel {
				return
			}

			var entry *tapApi.Entry
			err = json.Unmarshal(bytes, &entry)
			if err != nil {
				logger.Log.Debugf("error unmarshalling entry: %v", err.Error())
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

			entryStreamerSocketConnector.SendMetadata(socketId, metadata)
		}
	}

	go handleDataChannel(connection, data)
	go handleMetaChannel(connection, meta)

	connection.Query(query, data, meta)

	go func() {
		<-ctx.Done()
		data <- []byte(basenine.CloseChannel)
		meta <- []byte(basenine.CloseChannel)
		connection.Close()
	}()

	return cancel, nil
}
