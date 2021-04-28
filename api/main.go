package main

import (
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/inserter"
	"mizuserver/pkg/middleware"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/utils"
)



func main() {

	app := fiber.New()

	go inserter.StartReadingFiles("/tmp/up9hars")  // process to read files and insert to DB

	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.

	app.Static("/", "./site")

	//Simple route to know server is running
	app.Get("/echo", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	routes.WebSocketRoutes(app)
	routes.EntriesRoutes(app)
	routes.NotFoundRoute(app)

	utils.StartServer(app)
}
