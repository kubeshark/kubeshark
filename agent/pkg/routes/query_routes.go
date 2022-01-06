package routes

import (
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func QueryRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/query")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.POST("/validate", controllers.PostValidate)
}
