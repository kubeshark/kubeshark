package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

var (
	EntriesGetHandler       = controllers.GetEntries
	EntriesGetSingleHandler = controllers.GetEntry
)

// EntriesRoutes defines the group of har entries routes.
func EntriesRoutes(ginApp *gin.Engine) *gin.RouterGroup {
	routeGroup := ginApp.Group("/entries")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.GET("/", func(c *gin.Context) { EntriesGetHandler(c) })          // get entries (base/thin entries) and metadata
	routeGroup.GET("/:id", func(c *gin.Context) { EntriesGetSingleHandler(c) }) // get single (full) entry

	return routeGroup
}
