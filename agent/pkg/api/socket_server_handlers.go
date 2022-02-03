package api

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/agent/pkg/providers"
	"github.com/up9inc/mizu/agent/pkg/providers/tappers"
	"github.com/up9inc/mizu/agent/pkg/up9"

	tapApi "github.com/up9inc/mizu/tap/api"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

var browserClientSocketUUIDToTopic = make(map[int]string)
var socketListLock = sync.Mutex{}

type RoutesEventHandlers struct {
	EventHandlers
	SocketOutChannel chan<- *tapApi.OutputChannelItem
}

func init() {
	go up9.UpdateAnalyzeStatus(BroadcastToBrowserClients)
}

func (h *RoutesEventHandlers) WebSocketConnect(socketId int, isTapper bool, topic string) {
	if isTapper {
		logger.Log.Infof("Websocket event - Tapper connected, socket ID: %d", socketId)
		tappers.Connected()
	} else {
		logger.Log.Infof("Websocket event - Browser socket connected, socket ID: %d", socketId)
		socketListLock.Lock()
		browserClientSocketUUIDToTopic[socketId] = topic
		socketListLock.Unlock()
	}
}

func (h *RoutesEventHandlers) WebSocketDisconnect(socketId int, isTapper bool) {
	if isTapper {
		logger.Log.Infof("Websocket event - Tapper disconnected, socket ID:  %d", socketId)
		tappers.Disconnected()
	} else {
		logger.Log.Infof("Websocket event - Browser socket disconnected, socket ID:  %d", socketId)
		socketListLock.Lock()
		delete(browserClientSocketUUIDToTopic, socketId)
		socketListLock.Unlock()
	}
}

func BroadcastToBrowserClients(message []byte) {
	BroadcastToBrowserClientsByTopic(message, nil)
}

func BroadcastToBrowserClientsByTopic(message []byte, topic *string) {
	var t string
	if topic == nil {
		t = "nil"
	} else {
		t = *topic
	}

	logger.Log.Infof("Broadcasting message %s to browser clients of topic %s", string(message), t)
	for socketId, socketTopic := range browserClientSocketUUIDToTopic {
		if topic == nil || socketTopic == *topic {
			go func(socketId int) {
				if err := SendToSocket(socketId, message); err != nil {
					logger.Log.Error(err)
				}
			}(socketId)
		}
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
			logger.Log.Infof("received message of type %s", socketMessageBase.MessageType)
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
	resolvedNameObject := k8sResolver.Resolve(outboundLinkMessage.Data.DstIP)
	if resolvedNameObject != nil {
		outboundLinkMessage.Data.DstIP = resolvedNameObject.FullAddress
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
