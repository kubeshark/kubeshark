package routes

import (
	"mizuserver/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func QueryRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/query")

	routeGroup.POST("/validate", controllers.PostValidate)
}
