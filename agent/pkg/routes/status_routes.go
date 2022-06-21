package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/controllers"
)

func StatusRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/status")

	routeGroup.GET("/health", controllers.HealthCheck)

	routeGroup.POST("/tappedPods", controllers.PostTappedPods)
	routeGroup.POST("/tapperStatus", controllers.PostTapperStatus)
	routeGroup.GET("/connectedTappersCount", controllers.GetConnectedTappersCount)
	routeGroup.GET("/tap", controllers.GetTappingStatus)

	routeGroup.GET("/general", controllers.GetGeneralStats) // get general stats about entries in DB
	routeGroup.GET("/accumulative", controllers.GetAccumulativeStats)
	routeGroup.GET("/accumulativeTiming", controllers.GetAccumulativeStatsTiming)
	routeGroup.GET("/accumulativeBoth", controllers.GetAccumulativeStatsAll)

	routeGroup.GET("/resolving", controllers.GetCurrentResolvingInformation)
}
