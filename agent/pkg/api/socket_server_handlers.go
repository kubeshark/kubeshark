package api

import (
	"encoding/json"
	"fmt"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
	"mizuserver/pkg/models"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/up9"
	"sync"
)

var browserClientSocketUUIDs = make([]int, 0)
var socketListLock = sync.Mutex{}

type RoutesEventHandlers struct {
	routes.EventHandlers
	SocketHarOutChannel chan<- *tap.OutputChannelItem
}

func init() {
	go up9.UpdateAnalyzeStatus(broadcastToBrowserClients)
}

func (h *RoutesEventHandlers) WebSocketConnect(socketId int, isTapper bool) {
	if isTapper {
		rlog.Infof("Websocket Connection event - Tapper connected: %s", socketId)
	} else {
		rlog.Infof("Websocket Connection event - Browser socket connected: %s", socketId)
		socketListLock.Lock()
		browserClientSocketUUIDs = append(browserClientSocketUUIDs, socketId)
		socketListLock.Unlock()
	}
}

func (h *RoutesEventHandlers) WebSocketDisconnect(socketId int, isTapper bool) {
	if isTapper {
		rlog.Infof("Disconnection event - Tapper connected: %s", socketId)
	} else {
		rlog.Infof("Disconnection event - Browser socket connected: %s", socketId)
		socketListLock.Lock()
		removeSocketUUIDFromBrowserSlice(socketId)
		socketListLock.Unlock()
	}
}

func broadcastToBrowserClients(message []byte) {
	for _, socketId := range browserClientSocketUUIDs {
		go func(socketId int) {
			err := routes.SendToSocket(socketId, message)
			if err != nil {
				fmt.Printf("error sending message to socket id %d: %v", socketId, err)
			}
		}(socketId)

	}
}

func (h *RoutesEventHandlers) WebSocketMessage(_ int, message []byte) {
	var socketMessageBase shared.WebSocketMessageMetadata
	err := json.Unmarshal(message, &socketMessageBase)
	if err != nil {
		rlog.Infof("Could not unmarshal websocket message %v\n", err)
	} else {
		switch socketMessageBase.MessageType {
		case shared.WebSocketMessageTypeTappedEntry:
			var tappedEntryMessage models.WebSocketTappedEntryMessage
			err := json.Unmarshal(message, &tappedEntryMessage)
			if err != nil {
				rlog.Infof("Could not unmarshal message of message type %s %v\n", socketMessageBase.MessageType, err)
			} else {
				h.SocketHarOutChannel <- tappedEntryMessage.Data
			}
		case shared.WebSocketMessageTypeUpdateStatus:
			var statusMessage shared.WebSocketStatusMessage
			err := json.Unmarshal(message, &statusMessage)
			if err != nil {
				rlog.Infof("Could not unmarshal message of message type %s %v\n", socketMessageBase.MessageType, err)
			} else {
				providers.TapStatus.Pods = statusMessage.TappingStatus.Pods
				broadcastToBrowserClients(message)
			}
		case shared.WebsocketMessageTypeOutboundLink:
			var outboundLinkMessage models.WebsocketOutboundLinkMessage
			err := json.Unmarshal(message, &outboundLinkMessage)
			if err != nil {
				rlog.Infof("Could not unmarshal message of message type %s %v\n", socketMessageBase.MessageType, err)
			} else {
				handleTLSLink(outboundLinkMessage)
			}
		default:
			rlog.Infof("Received socket message of type %s for which no handlers are defined", socketMessageBase.MessageType)
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
		rlog.Errorf("Error marshaling outbound link message for broadcasting: %v", err)
	} else {
		fmt.Printf("Broadcasting outboundlink message %s\n", string(marshaledMessage))
		broadcastToBrowserClients(marshaledMessage)
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
