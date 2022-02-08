package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func InstallRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/install")

	routeGroup.GET("/isNeeded", controllers.IsSetupNecessary)
	routeGroup.POST("/admin", controllers.SetupAdminUser)
}
