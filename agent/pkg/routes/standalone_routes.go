package routes

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/controllers"
)

func StandaloneRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/standalone")

	routeGroup.POST("/updateConfig", controllers.UpdateConfig)
	routeGroup.GET("/config", controllers.GetConfig)
}
