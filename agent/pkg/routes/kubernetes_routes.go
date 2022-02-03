package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/controllers"
)

func KubernetesRoutes(app *gin.Engine) {
	routeGroup := app.Group("/kube")

	routeGroup.GET("/namespaces", controllers.GetNamespaces)
}
