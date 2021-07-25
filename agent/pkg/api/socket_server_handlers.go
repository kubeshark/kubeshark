package api

import (
	"encoding/json"
	"fmt"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/models"
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
				controllers.TapStatus = statusMessage.TappingStatus
				broadcastToBrowserClients(message)
			}
		default:
			rlog.Infof("Received socket message of type %s for which no handlers are defined", socketMessageBase.MessageType)
		}
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
