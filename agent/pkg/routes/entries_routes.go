package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

// EntriesRoutes defines the group of har entries routes.
func EntriesRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/entries")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.GET("/", controllers.GetEntries)  // get entries (base/thin entries) and metadata
	routeGroup.GET("/:id", controllers.GetEntry) // get single (full) entry
}
