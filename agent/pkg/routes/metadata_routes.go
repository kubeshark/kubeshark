package routes

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/controllers"
)

// MetadataRoutes defines the group of metadata routes.
func MetadataRoutes(app *gin.Engine) {
	routeGroup := app.Group("/metadata")

	routeGroup.GET("/version", controllers.GetVersion)
	routeGroup.GET("/health", controllers.HealthCheck)
}
