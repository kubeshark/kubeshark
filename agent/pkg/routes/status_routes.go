package routes

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/controllers"
)

func StatusRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/status")

	routeGroup.POST("/tappedPods", controllers.PostTappedPods)

	routeGroup.GET("/tappersCount", controllers.GetTappersCount)

	routeGroup.GET("/auth", controllers.GetAuthStatus)
}
