package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/agent/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

var (
	OASGetServersHandler    = controllers.GetOASServers
	OASGetAllSpecsHandler   = controllers.GetOASAllSpecs
	OASGetSingleSpecHandler = controllers.GetOASSpec
)

// OASRoutes methods to access OAS spec
func OASRoutes(ginApp *gin.Engine) *gin.RouterGroup {
	routeGroup := ginApp.Group("/oas")
	routeGroup.Use(middlewares.RequiresAuth())

	routeGroup.GET("/", func(c *gin.Context) { OASGetServersHandler(c) })       // list of servers in OAS map
	routeGroup.GET("/all", func(c *gin.Context) { OASGetAllSpecsHandler(c) })   // list of servers in OAS map
	routeGroup.GET("/:id", func(c *gin.Context) { OASGetSingleSpecHandler(c) }) // get OAS spec for given server

	return routeGroup
}
