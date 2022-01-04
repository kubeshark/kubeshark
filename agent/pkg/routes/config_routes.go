package routes

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/controllers"
)

func ConfigRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/config")

	routeGroup.POST("/tapConfig", controllers.PostTapConfig)
	routeGroup.GET("/tapConfig", controllers.GetTapConfig)
}
