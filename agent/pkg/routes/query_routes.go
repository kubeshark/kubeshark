package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

var (
	QueryPostValidateHandler = controllers.PostValidate
)

func QueryRoutes(ginApp *gin.Engine) *gin.RouterGroup {
	routeGroup := ginApp.Group("/query")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.POST("/validate", func(c *gin.Context) { QueryPostValidateHandler(c) })

	return routeGroup
}
