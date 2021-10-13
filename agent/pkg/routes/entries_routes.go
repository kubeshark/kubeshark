package routes

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/controllers"
)

// EntriesRoutes defines the group of har entries routes.
func EntriesRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/entries")

	routeGroup.GET("/", controllers.GetEntries)        // get entries (base/thin entries)
	routeGroup.GET("/:entryId", controllers.GetEntry) // get single (full) entry

	routeGroup.GET("/export", controllers.GetFullEntries)

	routeGroup.GET("/resetDB", controllers.DeleteAllEntries)     // get single (full) entry
}
