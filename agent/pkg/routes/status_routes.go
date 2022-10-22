package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/kubeshark/agent/pkg/controllers"
)

func StatusRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/status")

	routeGroup.GET("/health", controllers.HealthCheck)

	routeGroup.POST("/tappedPods", controllers.PostTappedPods)
	routeGroup.POST("/tapperStatus", controllers.PostTapperStatus)
	routeGroup.GET("/connectedTappersCount", controllers.GetConnectedTappersCount)
	routeGroup.GET("/tap", controllers.GetTappingStatus)

	routeGroup.GET("/general", controllers.GetGeneralStats)
	routeGroup.GET("/trafficStats", controllers.GetTrafficStats)

	routeGroup.GET("/resolving", controllers.GetCurrentResolvingInformation)
}
