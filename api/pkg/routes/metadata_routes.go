package routes

import (
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/controllers"
)

// MetadataRoutes defines the group of metadata routes.
func MetadataRoutes(fiberApp *fiber.App) {
	routeGroup := fiberApp.Group("/metadata")

	routeGroup.Get("/version", controllers.GetVersion)
}
