package routes

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/controllers"
)

func StandaloneRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/standalone")

	routeGroup.POST("/tapConfig", controllers.PostTapConfig)
	routeGroup.GET("/tapConfig", controllers.GetTapConfig)
}
