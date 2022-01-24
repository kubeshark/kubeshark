package routes

import (
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func ConfigRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/config")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.POST("/tapConfig", middlewares.RequiresAdmin(), controllers.PostTapConfig)
	routeGroup.GET("/tapConfig", controllers.GetTapConfig)
}
