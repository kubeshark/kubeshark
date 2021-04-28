package main

import (
	"fmt"
	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/inserter"
	"mizuserver/pkg/middleware"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/utils"
)

func main() {

	app := fiber.New()

	//	TODO: disabling this line for now (this should be as part of the MAIN
	go inserter.StartReadingFiles("/var/up9hars")
	//	TODO: to generate data
	// path := "/Users/roeegadot/Downloads/output2"
	// api.TestHarSavingFromFolder(path)

	middleware.FiberMiddleware(app) // Register Fiber's middleware for app. // TODO: Check if working

	app.Static("/", "./site")
	//Simple route to know server is running
	app.Get("/echo", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	// Multiple event handling supported
	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Connection event 1 - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Connection event 2 - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On message event
	ikisocket.On(ikisocket.EventMessage, func(ep *ikisocket.EventPayload) {

		ikisocket.Broadcast([]byte("bla"))
		fmt.Println(fmt.Sprintf("Message event - User: %s - Message: %s", ep.Kws.GetStringAttribute("user_id"), string(ep.Data)))
	})

	// On disconnect event
	ikisocket.On(ikisocket.EventDisconnect, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On close event
	// This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventClose, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		fmt.Println(fmt.Sprintf("Close event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On error event
	ikisocket.On(ikisocket.EventError, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Error event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
		// kws.Broadcast([]byte(fmt.Sprintf("New user connected: %s and UUID: %s", userId, kws.UUID)), true)
		// kws.Emit([]byte(fmt.Sprintf("Hello user with UUID: %s", kws.UUID)))
		fmt.Println("User connected")
	}))

	routes.EntriesRoutes(app)
	routes.NotFoundRoute(app)

	utils.StartServer(app)
}
