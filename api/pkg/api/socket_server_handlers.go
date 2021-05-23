package api

import (
	"encoding/json"
	"fmt"
	"github.com/antoniodipinto/ikisocket"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/controllers"
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
	fmt.Println("Received socket message")
	//TODO: singular flow for unmarshalling messages, tapper messages should use ControlSocketMessage (Which should be renamed)
	if ep.Kws.GetAttribute("is_tapper") == true && h.SocketHarOutChannel != nil{
		var tapOutput tap.OutputChannelItem
		err := json.Unmarshal(ep.Data, &tapOutput)
		if err != nil {
			fmt.Printf("Could not unmarshal message received from tapper websocket %v", err)
		} else {
			h.SocketHarOutChannel <- &tapOutput
		}
	} else {
		var controlMessage shared.ControlSocketMessage
		err := json.Unmarshal(ep.Data, &controlMessage)
		if err != nil {
			fmt.Printf("Could not unmarshal message received from tapper websocket %v\n", err)
		} else {
			if controlMessage.MessageType == shared.TAPPING_STATUS_MESSAGE_TYPE {
				tappingUpdateData, ok := controlMessage.Data.(*shared.TapStatus)
				if ok {
					controllers.TapStatus = tappingUpdateData
				} else {
					fmt.Printf("Could not cast tap status socket message")
				}
			} else {
				fmt.Printf("Received socket message of type %s for which no handlers are defined", controlMessage.MessageType)
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
