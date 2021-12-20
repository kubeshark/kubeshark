package api

import (
	"encoding/json"
	"fmt"
	"mizuserver/pkg/models"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/up9"
	"sync"

	tapApi "github.com/up9inc/mizu/tap/api"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

var browserClientSocketUUIDs = make([]int, 0)
var socketListLock = sync.Mutex{}

type RoutesEventHandlers struct {
	EventHandlers
	SocketOutChannel chan<- *tapApi.OutputChannelItem
}

func init() {
	go up9.UpdateAnalyzeStatus(BroadcastToBrowserClients)
}

func (h *RoutesEventHandlers) WebSocketConnect(socketId int, isTapper bool) {
	if isTapper {
		logger.Log.Infof("Websocket event - Tapper connected, socket ID: %d", socketId)
		providers.TapperAdded()
	} else {
		logger.Log.Infof("Websocket event - Browser socket connected, socket ID: %d", socketId)
		socketListLock.Lock()
		browserClientSocketUUIDs = append(browserClientSocketUUIDs, socketId)
		socketListLock.Unlock()
	}
}

func (h *RoutesEventHandlers) WebSocketDisconnect(socketId int, isTapper bool) {
	if isTapper {
		logger.Log.Infof("Websocket event - Tapper disconnected, socket ID:  %d", socketId)
		providers.TapperRemoved()
	} else {
		logger.Log.Infof("Websocket event - Browser socket disconnected, socket ID:  %d", socketId)
		socketListLock.Lock()
		removeSocketUUIDFromBrowserSlice(socketId)
		socketListLock.Unlock()
	}
}

func BroadcastToBrowserClients(message []byte) {
	for _, socketId := range browserClientSocketUUIDs {
		go func(socketId int) {
			err := SendToSocket(socketId, message)
			if err != nil {
				logger.Log.Errorf("error sending message to socket ID %d: %v", socketId, err)
			}
		}(socketId)
	}
}

func (h *RoutesEventHandlers) WebSocketMessage(_ int, message []byte) {
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
		case shared.WebsocketMessageTypeOutboundLink:
			var outboundLinkMessage models.WebsocketOutboundLinkMessage
			err := json.Unmarshal(message, &outboundLinkMessage)
			if err != nil {
				logger.Log.Infof("Could not unmarshal message of message type %s %v", socketMessageBase.MessageType, err)
			} else {
				handleTLSLink(outboundLinkMessage)
			}
		default:
			logger.Log.Infof("Received socket message of type %s for which no handlers are defined", socketMessageBase.MessageType)
		}
	}
}

func handleTLSLink(outboundLinkMessage models.WebsocketOutboundLinkMessage) {
	resolvedName := k8sResolver.Resolve(outboundLinkMessage.Data.DstIP)
	if resolvedName != "" {
		outboundLinkMessage.Data.DstIP = resolvedName
	} else if outboundLinkMessage.Data.SuggestedResolvedName != "" {
		outboundLinkMessage.Data.DstIP = outboundLinkMessage.Data.SuggestedResolvedName
	}
	cacheKey := fmt.Sprintf("%s -> %s:%d", outboundLinkMessage.Data.Src, outboundLinkMessage.Data.DstIP, outboundLinkMessage.Data.DstPort)
	_, isInCache := providers.RecentTLSLinks.Get(cacheKey)
	if isInCache {
		return
	} else {
		providers.RecentTLSLinks.SetDefault(cacheKey, outboundLinkMessage.Data)
	}
	marshaledMessage, err := json.Marshal(outboundLinkMessage)
	if err != nil {
		logger.Log.Errorf("Error marshaling outbound link message for broadcasting: %v", err)
	} else {
		logger.Log.Errorf("Broadcasting outboundlink message %s", string(marshaledMessage))
		BroadcastToBrowserClients(marshaledMessage)
	}
}

func removeSocketUUIDFromBrowserSlice(uuidToRemove int) {
	newUUIDSlice := make([]int, 0, len(browserClientSocketUUIDs))
	for _, uuid := range browserClientSocketUUIDs {
		if uuid != uuidToRemove {
			newUUIDSlice = append(newUUIDSlice, uuid)
		}
	}
	browserClientSocketUUIDs = newUUIDSlice
}
