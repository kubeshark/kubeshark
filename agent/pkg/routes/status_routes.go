package routes

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/controllers"
)

func StatusRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/status")

	routeGroup.GET("/health", controllers.HealthCheck)

	routeGroup.POST("/tappedPods", controllers.PostTappedPods)
	routeGroup.POST("/tapperStatus", controllers.PostTapperStatus)
	routeGroup.GET("/tappersCount", controllers.GetTappersCount)
	routeGroup.GET("/tap", controllers.GetTappingStatus)

	routeGroup.GET("/auth", controllers.GetAuthStatus)

	routeGroup.GET("/analyze", controllers.AnalyzeInformation)

	routeGroup.GET("/general", controllers.GetGeneralStats) // get general stats about entries in DB

	routeGroup.GET("/recentTLSLinks", controllers.GetRecentTLSLinks)

	routeGroup.GET("/resolving", controllers.GetCurrentResolvingInformation)
}
