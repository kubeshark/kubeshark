package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/controllers"
)

// ReplayRoutes defines the group of replay routes.
func ReplayRoutes(app *gin.Engine) {
	routeGroup := app.Group("/replay")

	routeGroup.POST("/", controllers.ReplayRequest)
	// this might be /replay/protocol? to create different functions for each type? but it can also be another parameter in the body
}
