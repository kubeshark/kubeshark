package routes

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/controllers"
)

// EntriesRoutes defines the group of har entries routes.
func EntriesRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/api")

	routeGroup.GET("/entries", controllers.GetEntries)        // get entries (base/thin entries)
	routeGroup.GET("/entries/:entryId", controllers.GetEntry) // get single (full) entry
	routeGroup.GET("/exportEntries", controllers.GetFullEntries)
	routeGroup.GET("/uploadEntries", controllers.UploadEntries)
	routeGroup.GET("/resolving", controllers.GetCurrentResolvingInformation)

	routeGroup.GET("/resetDB", controllers.DeleteAllEntries)     // get single (full) entry
	routeGroup.GET("/generalStats", controllers.GetGeneralStats) // get general stats about entries in DB

	routeGroup.GET("/tapStatus", controllers.GetTappingStatus) // get tapping status
	routeGroup.GET("/analyzeStatus", controllers.AnalyzeInformation)
	routeGroup.GET("/recentTLSLinks", controllers.GetRecentTLSLinks)
}
