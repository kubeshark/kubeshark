package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/controllers"
)

// OASRoutes methods to access OAS spec
func OASRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/oas")

	routeGroup.GET("/", controllers.GetOASServers)     // list of servers in OAS map
	routeGroup.GET("/all", controllers.GetOASAllSpecs) // list of servers in OAS map
	routeGroup.GET("/:id", controllers.GetOASSpec)     // get OAS spec for given server
}
