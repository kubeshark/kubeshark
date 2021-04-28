package main

import (
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/middleware"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/utils"
)

func main() {
	// TODO: to generate data
	//path := "/Users/roeegadot/Downloads/output2"
	//api.TestHarSavingFromFolder(path)

	// TODO: disabling this line for now (this should be as part of the MAIN
	// go inserter.StartReadingFiles("/var/up9hars")
	app := fiber.New()

	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.

	app.Static("/", "./site")

	// Simple route to know server is running
	app.Get("/echo", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	routes.EntriesRoutes(app)
	routes.NotFoundRoute(app)

	utils.StartServer(app)
}

