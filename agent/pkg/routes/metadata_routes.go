package routes

import (
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

// MetadataRoutes defines the group of metadata routes.
func MetadataRoutes(app *gin.Engine) {
	routeGroup := app.Group("/metadata")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.GET("/version", controllers.GetVersion)
}
