package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

// OASRoutes methods to access OAS spec
func OASRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/oas")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.GET("/", controllers.GetOASServers)     // list of servers in OAS map
	routeGroup.GET("/all", controllers.GetOASAllSpecs) // list of servers in OAS map
	routeGroup.GET("/:id", controllers.GetOASSpec)     // get OAS spec for given server
}
