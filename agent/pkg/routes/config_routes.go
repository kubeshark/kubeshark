package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func ConfigRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/config")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.POST("/tap", middlewares.RequiresAdmin(), controllers.PostTapConfig)
	routeGroup.GET("/tap", controllers.GetTapConfig)
}
