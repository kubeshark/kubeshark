package main

import (
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/inserter"
	"mizuserver/pkg/middleware"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/tap"
	"mizuserver/pkg/utils"
)

func main() {

	harOutputChannel := tap.StartPassiveTapper()

	app := fiber.New()

	// process to read files / channel and insert to DB
	go inserter.StartReadingFiles(harOutputChannel, tap.HarOutputDir)


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
