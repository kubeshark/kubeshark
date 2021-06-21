package routes

import (
	"github.com/gofiber/fiber/v2"
	"mizuserver/pkg/controllers"
)

// EntriesRoutes func for describe group of public routes.
func EntriesRoutes(fiberApp *fiber.App) {
	routeGroup := fiberApp.Group("/api")

	routeGroup.Get("/entries", controllers.GetEntries)        // get entries (base/thin entries)
	routeGroup.Get("/entries/:entryId", controllers.GetEntry) // get single (full) entry
	routeGroup.Get("/exportEntries", controllers.GetFullEntries)


	routeGroup.Get("/har", controllers.GetHARs)

	routeGroup.Get("/resetDB", controllers.DeleteAllEntries)     // get single (full) entry
	routeGroup.Get("/generalStats", controllers.GetGeneralStats) // get general stats about entries in DB

	routeGroup.Get("/tapStatus", controllers.GetTappingStatus) // get tapping status
}
