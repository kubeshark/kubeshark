package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"

	"github.com/gin-gonic/gin"
)

// MetadataRoutes defines the group of metadata routes.
func MetadataRoutes(app *gin.Engine) {
	routeGroup := app.Group("/metadata")

	routeGroup.GET("/version", controllers.GetVersion)
}
