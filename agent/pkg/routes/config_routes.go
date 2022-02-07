package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

var (
	ConfigPostTapConfigHandler = controllers.PostTapConfig
	ConfigGetTapConfigHandler  = controllers.GetTapConfig
)

func ConfigRoutes(ginApp *gin.Engine) *gin.RouterGroup {
	routeGroup := ginApp.Group("/config")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.POST("/tap", middlewares.RequiresAdmin(), func(c *gin.Context) { ConfigPostTapConfigHandler(c) })
	routeGroup.GET("/tap", func(c *gin.Context) { ConfigGetTapConfigHandler(c) })

	return routeGroup
}
