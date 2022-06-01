package api

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/dependency"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/agent/pkg/providers/tappedPods"
	"github.com/up9inc/mizu/agent/pkg/providers/tappers"

	tapApi "github.com/up9inc/mizu/tap/api"

	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
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
	for socketId := range browserClients {
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
		HandleTapperIncomingMessage(message, h.SocketOutChannel, BroadcastToBrowserClients)
	} else {
		// we initiate the basenine stream after the first websocket message we receive (it contains the entry query), we then store a cancelfunc to later cancel this stream
		if browserClients[socketId] != nil && browserClients[socketId].dataStreamCancelFunc == nil {
			var params WebSocketParams
			if err := json.Unmarshal(message, &params); err != nil {
				logger.Log.Errorf("Error: %v", socketId, err)
				return
			}

			entriesStreamer := dependency.GetInstance(dependency.EntriesSocketStreamer).(EntryStreamer)
			ctx, cancelFunc := context.WithCancel(context.Background())
			err := entriesStreamer.Get(ctx, socketId, &params)

			if err != nil {
				logger.Log.Errorf("error initializing basenine stream for browser socket %d %+v", socketId, err)
				cancelFunc()
			} else {
				browserClients[socketId].dataStreamCancelFunc = cancelFunc
			}
		}
	}
}

func HandleTapperIncomingMessage(message []byte, socketOutChannel chan<- *tapApi.OutputChannelItem, broadcastMessageFunc func([]byte)) {
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
				socketOutChannel <- tappedEntryMessage.Data
			}
		case shared.WebSocketMessageTypeUpdateStatus:
			var statusMessage shared.WebSocketStatusMessage
			err := json.Unmarshal(message, &statusMessage)
			if err != nil {
				logger.Log.Infof("Could not unmarshal message of message type %s %v", socketMessageBase.MessageType, err)
			} else {
				broadcastMessageFunc(message)
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
