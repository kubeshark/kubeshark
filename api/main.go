package main

import (
	"flag"
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/inserter"
	"mizuserver/pkg/middleware"
	"mizuserver/pkg/routes"
	"mizuserver/pkg/utils"
)



func main() {

	app := fiber.New()

	var harDir = flag.String("hardir", "input", "Directory in which we read har files from")
	go inserter.StartReadingFiles(*harDir)  // process to read files and insert to DB


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
