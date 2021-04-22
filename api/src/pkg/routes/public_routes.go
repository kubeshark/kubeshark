package routes

import (
	"mizuserver/src/pkg/controllers"

	"github.com/gofiber/fiber/v2"
)

// PublicRoutes func for describe group of public routes.
func PublicRoutes(a *fiber.App) {
	// Create routes group.
	route := a.Group("/api")

	// Routes for GET method:
	route.Get("/entries", controllers.GetBooks)   // get list of all books
	route.Get("/entries/:id", controllers.GetBook) // get one book by ID

}
