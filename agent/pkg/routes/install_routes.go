package routes

import (
	"mizuserver/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func InstallRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/install")

	routeGroup.GET("/isNeeded", controllers.IsSetupNecessary)
}
