package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

var (
	ServiceMapGetStatus gin.HandlerFunc
	ServiceMapGet       gin.HandlerFunc
	ServiceMapReset     gin.HandlerFunc
)

func ServiceMapRoutes(ginApp *gin.Engine) *gin.RouterGroup {
	routeGroup := ginApp.Group("/servicemap")
	routeGroup.Use(middlewares.RequiresAuth())

	controller := controllers.NewServiceMapController()

	ServiceMapGetStatus = controller.Status
	ServiceMapGet = controller.Get
	ServiceMapReset = controller.Reset

	routeGroup.GET("/status", func(c *gin.Context) { ServiceMapGetStatus(c) })
	routeGroup.GET("/get", func(c *gin.Context) { ServiceMapGet(c) })
	routeGroup.GET("/reset", func(c *gin.Context) { ServiceMapReset(c) })

	return routeGroup
}
