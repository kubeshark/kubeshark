package routes

import (
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

// EntriesRoutes defines the group of har entries routes.
func EntriesRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/entries")

	routeGroup.GET("/", controllers.GetEntries)  // get entries (base/thin entries) and metadata
	routeGroup.GET("/:id", controllers.GetEntry) // get single (full) entry

	routeGroup.Use(middlewares.RequiresAuth)
}
