package api

import (
	"encoding/json"
	"github.com/antoniodipinto/ikisocket"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/models"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/up9"
)

var browserClientSocketUUIDs = make([]string, 0)

type RoutesEventHandlers struct {
	routes.EventHandlers
	SocketHarOutChannel chan<- *tap.OutputChannelItem
}

func init() {
	go up9.UpdateAnalyzeStatus(broadcastToBrowserClients)
}

func (h *RoutesEventHandlers) WebSocketConnect(ep *ikisocket.EventPayload) {
	if ep.Kws.GetAttribute("is_tapper") == true {
		rlog.Infof("Websocket Connection event - Tapper connected: %s", ep.SocketUUID)
	} else {
		rlog.Infof("Websocket Connection event - Browser socket connected: %s", ep.SocketUUID)
		browserClientSocketUUIDs = append(browserClientSocketUUIDs, ep.SocketUUID)
	}
}

func (h *RoutesEventHandlers) WebSocketDisconnect(ep *ikisocket.EventPayload) {
	if ep.Kws.GetAttribute("is_tapper") == true {
		rlog.Infof("Disconnection event - Tapper connected: %s", ep.SocketUUID)
	} else {
		rlog.Infof("Disconnection event - Browser socket connected: %s", ep.SocketUUID)
		removeSocketUUIDFromBrowserSlice(ep.SocketUUID)
	}
}

func broadcastToBrowserClients(message []byte) {
	ikisocket.EmitToList(browserClientSocketUUIDs, message)
}

func (h *RoutesEventHandlers) WebSocketClose(ep *ikisocket.EventPayload) {
	if ep.Kws.GetAttribute("is_tapper") == true {
		rlog.Infof("Websocket Close event - Tapper connected: %s", ep.SocketUUID)
	} else {
		rlog.Infof("Websocket  Close event - Browser socket connected: %s", ep.SocketUUID)
		removeSocketUUIDFromBrowserSlice(ep.SocketUUID)
	}
}

func (h *RoutesEventHandlers) WebSocketError(ep *ikisocket.EventPayload) {
	rlog.Infof("Socket error - Socket uuid : %s %v", ep.SocketUUID, ep.Error)
}

func (h *RoutesEventHandlers) WebSocketMessage(ep *ikisocket.EventPayload) {
	var socketMessageBase shared.WebSocketMessageMetadata
	err := json.Unmarshal(ep.Data, &socketMessageBase)
	if err != nil {
		rlog.Infof("Could not unmarshal websocket message %v\n", err)
	} else {
		switch socketMessageBase.MessageType {
		case shared.WebSocketMessageTypeTappedEntry:
			var tappedEntryMessage models.WebSocketTappedEntryMessage
			err := json.Unmarshal(ep.Data, &tappedEntryMessage)
			if err != nil {
				rlog.Infof("Could not unmarshal message of message type %s %v\n", socketMessageBase.MessageType, err)
			} else {
				h.SocketHarOutChannel <- tappedEntryMessage.Data
			}
		case shared.WebSocketMessageTypeUpdateStatus:
			var statusMessage shared.WebSocketStatusMessage
			err := json.Unmarshal(ep.Data, &statusMessage)
			if err != nil {
				rlog.Infof("Could not unmarshal message of message type %s %v\n", socketMessageBase.MessageType, err)
			} else {
				controllers.TapStatus = statusMessage.TappingStatus
				broadcastToBrowserClients(ep.Data)
			}
		default:
			rlog.Infof("Received socket message of type %s for which no handlers are defined", socketMessageBase.MessageType)
		}
	}
}

func removeSocketUUIDFromBrowserSlice(uuidToRemove string) {
	newUUIDSlice := make([]string, 0, len(browserClientSocketUUIDs))
	for _, uuid := range browserClientSocketUUIDs {
		if uuid != uuidToRemove {
			newUUIDSlice = append(newUUIDSlice, uuid)
		}
	}
	browserClientSocketUUIDs = newUUIDSlice
}
