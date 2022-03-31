package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/agent/pkg/providers/tappedPods"
	"github.com/up9inc/mizu/agent/pkg/providers/tappers"
	"github.com/up9inc/mizu/agent/pkg/up9"

	tapApi "github.com/up9inc/mizu/tap/api"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

type BrowserClient struct {
	dataStreamCancelFunc context.CancelFunc
}

var browserClients = make(map[int]*BrowserClient, 0)
var tapperClientSocketUUIDs = make([]int, 0)
var socketListLock = sync.Mutex{}

type RoutesEventHandlers struct {
	EventHandlers
	SocketOutChannel chan<- *tapApi.OutputChannelItem
}

func init() {
	go up9.UpdateAnalyzeStatus(BroadcastToBrowserClients)
}

func (h *RoutesEventHandlers) WebSocketConnect(_ *gin.Context, socketId int, isTapper bool) {
	if isTapper {
		logger.Log.Infof("Websocket event - Tapper connected, socket ID: %d", socketId)
		tappers.Connected()

		socketListLock.Lock()
		tapperClientSocketUUIDs = append(tapperClientSocketUUIDs, socketId)
		socketListLock.Unlock()

		nodeToTappedPodMap := tappedPods.GetNodeToTappedPodMap()
		SendTappedPods(socketId, nodeToTappedPodMap)
	} else {
		logger.Log.Infof("Websocket event - Browser socket connected, socket ID: %d", socketId)

		socketListLock.Lock()
		browserClients[socketId] = &BrowserClient{}
		socketListLock.Unlock()

		BroadcastTappedPodsStatus()
	}
}

func (h *RoutesEventHandlers) WebSocketDisconnect(socketId int, isTapper bool) {
	if isTapper {
		logger.Log.Infof("Websocket event - Tapper disconnected, socket ID:  %d", socketId)
		tappers.Disconnected()

		socketListLock.Lock()
		removeSocketUUIDFromTapperSlice(socketId)
		socketListLock.Unlock()
	} else {
		logger.Log.Infof("Websocket event - Browser socket disconnected, socket ID:  %d", socketId)
		socketListLock.Lock()
		if browserClients[socketId] != nil && browserClients[socketId].dataStreamCancelFunc != nil {
			browserClients[socketId].dataStreamCancelFunc()
		}
		delete(browserClients, socketId)
		socketListLock.Unlock()
	}
}

func BroadcastToBrowserClients(message []byte) {
	for socketId, _ := range browserClients {
		go func(socketId int) {
			if err := SendToSocket(socketId, message); err != nil {
				logger.Log.Error(err)
			}
		}(socketId)
	}
}

func BroadcastToTapperClients(message []byte) {
	for _, socketId := range tapperClientSocketUUIDs {
		go func(socketId int) {
			if err := SendToSocket(socketId, message); err != nil {
				logger.Log.Error(err)
			}
		}(socketId)
	}
}

func (h *RoutesEventHandlers) WebSocketMessage(socketId int, isTapper bool, message []byte) {
	if isTapper {
		h.handleTapperIncomingMessage(message)
	} else {
		// we initiate the basenine stream after the first websocket message we receive (it contains the entry query), we then store a cancelfunc to later cancel this stream
		if browserClients[socketId] != nil && browserClients[socketId].dataStreamCancelFunc == nil {
			cancelFunc, err := startStreamingBasenineEntriesToSocket(socketId, message)
			if err != nil {
				logger.Log.Errorf("error initializing basenine stream for browser socket %d %+v", socketId, err)
			} else {
				browserClients[socketId].dataStreamCancelFunc = cancelFunc
			}
		}
	}
}

func startStreamingBasenineEntriesToSocket(socketId int, message []byte) (context.CancelFunc, error) {
	var params WebSocketParams
	if err := json.Unmarshal(message, &params); err != nil {
		logger.Log.Errorf("error unmarshalling parameters: %v", socketId, err)
		return nil, err
	}

	var connection *basenine.Connection

	connection, err := basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
	if err != nil {
		logger.Log.Errorf("failed to establish a connection to Basenine: %v", err)
		socketCleanup(socketId, connectedWebsockets[socketId])
		return nil, err
	}

	data := make(chan []byte)
	meta := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())

	query := params.Query
	err = basenine.Validate(shared.BasenineHost, shared.BaseninePort, query)
	if err != nil {
		toastBytes, _ := models.CreateWebsocketToastMessage(&models.ToastMessage{
			Type:      "error",
			AutoClose: 5000,
			Text:      fmt.Sprintf("syntax error: %s", err.Error()),
		})
		if err := SendToSocket(socketId, toastBytes); err != nil {
			logger.Log.Error(err)
		}
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

			metadataBytes, _ := models.CreateWebsocketQueryMetadataMessage(metadata)
			if err := SendToSocket(socketId, metadataBytes); err != nil {
				logger.Log.Error(err)
			}
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

func (h *RoutesEventHandlers) handleTapperIncomingMessage(message []byte) {
	var socketMessageBase shared.WebSocketMessageMetadata
	err := json.Unmarshal(message, &socketMessageBase)
	if err != nil {
		logger.Log.Infof("Could not unmarshal websocket message %v", err)
	} else {
		switch socketMessageBase.MessageType {
		case shared.WebSocketMessageTypeTappedEntry:
			var tappedEntryMessage models.WebSocketTappedEntryMessage
			err := json.Unmarshal(message, &tappedEntryMessage)
			if err != nil {
				logger.Log.Infof("Could not unmarshal message of message type %s %v", socketMessageBase.MessageType, err)
			} else {
				// NOTE: This is where the message comes back from the intermediate WebSocket to code.
				h.SocketOutChannel <- tappedEntryMessage.Data
			}
		case shared.WebSocketMessageTypeUpdateStatus:
			var statusMessage shared.WebSocketStatusMessage
			err := json.Unmarshal(message, &statusMessage)
			if err != nil {
				logger.Log.Infof("Could not unmarshal message of message type %s %v", socketMessageBase.MessageType, err)
			} else {
				BroadcastToBrowserClients(message)
			}
		default:
			logger.Log.Infof("Received socket message of type %s for which no handlers are defined", socketMessageBase.MessageType)
		}
	}
}

func removeSocketUUIDFromTapperSlice(uuidToRemove int) {
	newUUIDSlice := make([]int, 0, len(tapperClientSocketUUIDs))
	for _, uuid := range tapperClientSocketUUIDs {
		if uuid != uuidToRemove {
			newUUIDSlice = append(newUUIDSlice, uuid)
		}
	}
	tapperClientSocketUUIDs = newUUIDSlice
}
