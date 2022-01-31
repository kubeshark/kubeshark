package routes

import (
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func ConfigRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/config")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.POST("/tap", middlewares.RequiresAdmin(), controllers.PostTapConfig)
	routeGroup.GET("/tap", controllers.GetTapConfig)
}
