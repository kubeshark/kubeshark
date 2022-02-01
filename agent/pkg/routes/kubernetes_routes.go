package routes

import (
	"mizuserver/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func KubernetesRoutes(app *gin.Engine) {
	routeGroup := app.Group("/kube")

	routeGroup.GET("/namespaces", controllers.GetNamespaces)
}
