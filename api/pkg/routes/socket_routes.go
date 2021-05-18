package socket

import (
	"fmt"
	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
)

var BrowserClientSocketUUIDs = make([]string, 0)

func webSocketConnect(ep *ikisocket.EventPayload) {
	if ep.Kws.GetAttribute("is_tapper") == true {
		fmt.Println("Connection event - Tapper connected")
	} else {
		fmt.Println(fmt.Sprintf("Connection event 1 - User: %s", ep.Kws.GetStringAttribute("user_id")))
		BrowserClientSocketUUIDs = append(BrowserClientSocketUUIDs, ep.SocketUUID)
	}

}

func webSocketDisconnect(ep *ikisocket.EventPayload) {
	//fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketClose(ep *ikisocket.EventPayload) {
	//fmt.Println(fmt.Sprintf("Close event - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketError(ep *ikisocket.EventPayload) {
	//fmt.Println(fmt.Sprintf("Error event - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketMessage(ep *ikisocket.EventPayload) {
	fmt.Println("Web socket message")

	// fmt.Println(fmt.Sprintf("Message event - User: %s - Message: %s", ep.Kws.GetStringAttribute("user_id"), string(ep.Data)))
}

func WebSocketRoutes(app *fiber.App) {

	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
		kws.Broadcast([]byte("hello ws"), true)
		kws.SetAttribute("user_id", kws.UUID)
	}))

	app.Get("/wsTapper", ikisocket.New(func(kws *ikisocket.Websocket) {
		kws.Broadcast([]byte("hello wsTapper"), true)
		//tapper clients are handled differently, they don't need to receive new message broadcasts
		kws.SetAttribute("is_tapper", true)
	}))

	ikisocket.On(ikisocket.EventMessage, webSocketMessage)
	ikisocket.On(ikisocket.EventConnect, webSocketConnect)
	ikisocket.On(ikisocket.EventDisconnect, webSocketDisconnect)
	ikisocket.On(ikisocket.EventClose, webSocketClose) // This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventError, webSocketError) // On error event
}
