package routes

import (
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/controllers"
)

// EntriesRoutes func for describe group of public routes.
func MetadataRoutes(fiberApp *fiber.App) {
	routeGroup := fiberApp.Group("/metadata")

	routeGroup.Get("/version", controllers.GetVersion)
}
