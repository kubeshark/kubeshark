package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/kubeshark/agent/pkg/controllers"
)

func ServiceMapRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/servicemap")

	controller := controllers.NewServiceMapController()

	routeGroup.GET("/status", controller.Status)
	routeGroup.GET("/get", controller.Get)
	routeGroup.GET("/reset", controller.Reset)
}
