package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func WorkspaceRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/workspace")

	routeGroup.GET("/:workspaceId", middlewares.RequiresAdmin(), controllers.GetWorkspace)
	routeGroup.PUT("/:workspaceId", middlewares.RequiresAdmin(), controllers.UpdateWorkspace)
	routeGroup.DELETE("/:workspaceId", middlewares.RequiresAdmin(), controllers.DeleteWorkspace)

	routeGroup.POST("/", middlewares.RequiresAdmin(), controllers.CreateWorkspace)
	routeGroup.GET("/", middlewares.RequiresAdmin(), controllers.ListWorkspace)
}
