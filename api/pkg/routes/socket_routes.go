package routes

import (
	"fmt"
	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
)

func webSocketConnect(ep *ikisocket.EventPayload) {
	fmt.Println(fmt.Sprintf("Connection event 1 - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketDisconnect(ep *ikisocket.EventPayload) {
	fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketClose(ep *ikisocket.EventPayload) {
	fmt.Println(fmt.Sprintf("Close event - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketError(ep *ikisocket.EventPayload) {
	fmt.Println(fmt.Sprintf("Error event - User: %s", ep.Kws.GetStringAttribute("user_id")))
}

func webSocketMessage(ep *ikisocket.EventPayload) {
	fmt.Println("Web socket message")
	// fmt.Println(fmt.Sprintf("Message event - User: %s - Message: %s", ep.Kws.GetStringAttribute("user_id"), string(ep.Data)))
}

func WebSocketRoutes(app *fiber.App) {

	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
		// kws.Broadcast([]byte(fmt.Sprintf("New user connected: %s and UUID: %s", userId, kws.UUID)), true)
		// kws.Emit([]byte(fmt.Sprintf("Hello user with UUID: %s", kws.UUID)))
		kws.SetAttribute("user_id", kws.UUID)
	}))

	ikisocket.On(ikisocket.EventMessage, webSocketMessage)
	ikisocket.On(ikisocket.EventConnect, webSocketConnect)
	ikisocket.On(ikisocket.EventDisconnect, webSocketDisconnect)
	ikisocket.On(ikisocket.EventClose, webSocketClose) // This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventError, webSocketError) // On error event
}
