package api

import (
	"encoding/json"
	"fmt"
	"github.com/antoniodipinto/ikisocket"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/models"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/tap"
)

var browserClientSocketUUIDs = make([]string, 0)

type RoutesEventHandlers struct {
	routes.EventHandlers
	SocketHarOutChannel chan<- *tap.OutputChannelItem
}


func (h *RoutesEventHandlers) WebSocketConnect(ep *ikisocket.EventPayload) {
	if ep.Kws.GetAttribute("is_tapper") == true {
		fmt.Println(fmt.Sprintf("Websocket Connection event - Tapper connected: %s", ep.SocketUUID))
	} else {
		fmt.Println(fmt.Sprintf("Websocket Connection event - Browser socket connected: %s", ep.SocketUUID))
		browserClientSocketUUIDs = append(browserClientSocketUUIDs, ep.SocketUUID)
	}
}

func (h *RoutesEventHandlers) WebSocketDisconnect(ep *ikisocket.EventPayload) {
	if ep.Kws.GetAttribute("is_tapper") == true {
		fmt.Println(fmt.Sprintf("Disconnection event - Tapper connected: %s", ep.SocketUUID))
	} else {
		fmt.Println(fmt.Sprintf("Disconnection event - Browser socket connected: %s", ep.SocketUUID))
		removeSocketUUIDFromBrowserSlice(ep.SocketUUID)
	}
}

func broadcastToBrowserClients(message []byte) {
	ikisocket.EmitToList(browserClientSocketUUIDs, message)
}

func (h *RoutesEventHandlers) WebSocketClose(ep *ikisocket.EventPayload) {
	if ep.Kws.GetAttribute("is_tapper") == true {
		fmt.Println(fmt.Sprintf("Websocket Close event - Tapper connected: %s", ep.SocketUUID))
	} else {
		fmt.Println(fmt.Sprintf("Websocket  Close event - Browser socket connected: %s", ep.SocketUUID))
		removeSocketUUIDFromBrowserSlice(ep.SocketUUID)
	}
}

func (h *RoutesEventHandlers) WebSocketError(ep *ikisocket.EventPayload) {
	fmt.Println(fmt.Sprintf("Socket error - Socket uuid : %s %v", ep.SocketUUID, ep.Error))
}

func (h *RoutesEventHandlers) WebSocketMessage(ep *ikisocket.EventPayload) {
	var socketMessageBase shared.WebSocketMessageMetadata
	err := json.Unmarshal(ep.Data, &socketMessageBase)
	if err != nil {
		fmt.Printf("Could not unmarshal websocket message %v\n", err)
	} else {
		switch socketMessageBase.MessageType {
		case shared.WebSocketMessageTypeTappedEntry:
			var tappedEntryMessage models.WebSocketTappedEntryMessage
			err := json.Unmarshal(ep.Data, &tappedEntryMessage)
			if err != nil {
				fmt.Printf("Could not unmarshal message of message type %s %v", socketMessageBase.MessageType, err)
			} else {
				h.SocketHarOutChannel <- tappedEntryMessage.Data
			}
		case shared.WebSocketMessageTypeUpdateStatus:
			var statusMessage shared.WebSocketStatusMessage
			err := json.Unmarshal(ep.Data, &statusMessage)
			if err != nil {
				fmt.Printf("Could not unmarshal message of message type %s %v", socketMessageBase.MessageType, err)
			} else {
				controllers.TapStatus = statusMessage.TappingStatus
				broadcastToBrowserClients(ep.Data)
			}
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
