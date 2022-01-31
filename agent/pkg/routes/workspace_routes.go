package routes

import (
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func WorkspaceRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/workspace")

	routeGroup.GET("/:workspaceId", middlewares.RequiresAdmin(), controllers.GetWorkspace)
	routeGroup.PUT("/:workspaceId", middlewares.RequiresAdmin(), controllers.UpdateWorkspace)
	routeGroup.POST("/", middlewares.RequiresAdmin(), controllers.CreateWorkspace)
	routeGroup.GET("/", middlewares.RequiresAdmin(), controllers.ListWorkspace)
}
