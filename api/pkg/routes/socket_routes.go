package routes

import (
	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
)

type EventHandlers interface {
	WebSocketConnect(ep *ikisocket.EventPayload)
	WebSocketDisconnect(ep *ikisocket.EventPayload)
	WebSocketClose(ep *ikisocket.EventPayload)
	WebSocketError(ep *ikisocket.EventPayload)
	WebSocketMessage(ep *ikisocket.EventPayload)
}

func WebSocketRoutes(app *fiber.App, eventHandlers EventHandlers) {
	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
		kws.SetAttribute("is_tapper", false)
	}))

	app.Get("/wsTapper", ikisocket.New(func(kws *ikisocket.Websocket) {
		// Tapper clients are handled differently, they don't need to receive new message broadcasts.
		kws.SetAttribute("is_tapper", true)
	}))

	ikisocket.On(ikisocket.EventMessage, eventHandlers.WebSocketMessage)
	ikisocket.On(ikisocket.EventConnect, eventHandlers.WebSocketConnect)
	ikisocket.On(ikisocket.EventDisconnect, eventHandlers.WebSocketDisconnect)
	ikisocket.On(ikisocket.EventClose, eventHandlers.WebSocketClose) // This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventError, eventHandlers.WebSocketError) // On error event
}
