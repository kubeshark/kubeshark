package routes

import (
	"mizuserver/src/pkg/controllers"

	"github.com/gofiber/fiber/v2"
)

// PublicRoutes func for describe group of public routes.
func PublicRoutes(a *fiber.App) {
	controllers.GenerateData()

	// Create routes group.
	route := a.Group("/api")

	// Routes for GET method:
	route.Get("/entries", controllers.GetEntries)        // get list of all books
	route.Get("/entries/:entryId", controllers.GetEntry) // get one book by ID

}
