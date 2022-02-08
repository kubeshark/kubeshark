package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"

	"github.com/gin-gonic/gin"
)

var (
	MetadataGetVersionHandler = controllers.GetVersion
)

// MetadataRoutes defines the group of metadata routes.
func MetadataRoutes(app *gin.Engine) *gin.RouterGroup {
	routeGroup := app.Group("/metadata")

	routeGroup.GET("/version", func(c *gin.Context) { MetadataGetVersionHandler(c) })

	return routeGroup
}
