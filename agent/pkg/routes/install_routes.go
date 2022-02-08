package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"

	"github.com/gin-gonic/gin"
)

var (
	InstallGetIsNeededHandler = controllers.IsSetupNecessary
	InstallPostAdminHandler   = controllers.SetupAdminUser
)

func InstallRoutes(ginApp *gin.Engine) *gin.RouterGroup {
	routeGroup := ginApp.Group("/install")

	routeGroup.GET("/isNeeded", func(c *gin.Context) { InstallGetIsNeededHandler(c) })
	routeGroup.POST("/admin", func(c *gin.Context) { InstallPostAdminHandler(c) })

	return routeGroup
}
