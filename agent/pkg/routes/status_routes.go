package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

var (
	StatusGetHealthCheck                 = controllers.HealthCheck
	StatusPostTappedPods                 = controllers.PostTappedPods
	StatusPostTapperStatus               = controllers.PostTapperStatus
	StatusGetConnectedTappersCount       = controllers.GetConnectedTappersCount
	StatusGetTappingStatus               = controllers.GetTappingStatus
	StatusGetAuthStatus                  = controllers.GetAuthStatus
	StatusGetAnalyzeInformation          = controllers.AnalyzeInformation
	StatusGetGeneralStats                = controllers.GetGeneralStats
	StatusGetRecentTLSLinks              = controllers.GetRecentTLSLinks
	StatusGetCurrentResolvingInformation = controllers.GetCurrentResolvingInformation
)

func StatusRoutes(ginApp *gin.Engine) *gin.RouterGroup {
	routeGroup := ginApp.Group("/status")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.GET("/health", func(c *gin.Context) { StatusGetHealthCheck(c) })

	routeGroup.POST("/tappedPods", func(c *gin.Context) { StatusPostTappedPods(c) })
	routeGroup.POST("/tapperStatus", func(c *gin.Context) { StatusPostTapperStatus(c) })
	routeGroup.GET("/connectedTappersCount", func(c *gin.Context) { StatusGetConnectedTappersCount(c) })
	routeGroup.GET("/tap", func(c *gin.Context) { StatusGetTappingStatus(c) })

	routeGroup.GET("/auth", func(c *gin.Context) { StatusGetAuthStatus(c) })

	routeGroup.GET("/analyze", func(c *gin.Context) { StatusGetAnalyzeInformation(c) })

	routeGroup.GET("/general", func(c *gin.Context) { StatusGetGeneralStats(c) }) // get general stats about entries in DB

	routeGroup.GET("/recentTLSLinks", func(c *gin.Context) { StatusGetRecentTLSLinks(c) })

	routeGroup.GET("/resolving", func(c *gin.Context) { StatusGetCurrentResolvingInformation(c) })

	return routeGroup
}
