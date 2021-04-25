package main

import (
	"github.com/gofiber/fiber/v2"
	api "mizuserver"
	"mizuserver/src/pkg/middleware"
	"mizuserver/src/pkg/routes"
	"mizuserver/src/pkg/utils"
)

func main() {
	// TODO: to generate data
	path := "/Users/roeegadot/Downloads/output2"
	api.TestHarSavingFromFolder(path)

	app := fiber.New()

	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.

	// Simple route to know server is running
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	routes.PublicRoutes(app)
	routes.NotFoundRoute(app)

	utils.StartServer(app)
}

