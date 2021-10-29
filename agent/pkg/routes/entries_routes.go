package routes

import (
	"mizuserver/pkg/controllers"

	"github.com/gin-gonic/gin"
)

// EntriesRoutes defines the group of har entries routes.
func EntriesRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/entries")

	routeGroup.GET("/:id", controllers.GetEntry) // get single (full) entry
}
