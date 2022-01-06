package routes

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"
)

func ConfigRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/config")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.POST("/tapConfig", controllers.PostTapConfig)
	routeGroup.GET("/tapConfig", controllers.GetTapConfig)
}
