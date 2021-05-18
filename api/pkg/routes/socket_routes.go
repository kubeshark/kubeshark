package routes

import (
	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/api"
)

func WebSocketRoutes(app *fiber.App) {

	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
		kws.SetAttribute("is_tapper", false)
	}))

	app.Get("/wsTapper", ikisocket.New(func(kws *ikisocket.Websocket) {
		//tapper clients are handled differently, they don't need to receive new message broadcasts
		kws.SetAttribute("is_tapper", true)
	}))

	ikisocket.On(ikisocket.EventMessage, api.WebSocketMessage)
	ikisocket.On(ikisocket.EventConnect, api.WebSocketConnect)
	ikisocket.On(ikisocket.EventDisconnect, api.WebSocketDisconnect)
	ikisocket.On(ikisocket.EventClose, api.WebSocketClose) // This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventError, api.WebSocketError) // On error event
}
