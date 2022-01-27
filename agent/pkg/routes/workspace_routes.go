package routes

import (
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func WorkspaceRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/user")

	routeGroup.GET("/:userId", middlewares.RequiresAdmin(), controllers.GetUser)

	routeGroup.GET("/whoAmI", middlewares.RequiresAuth(), controllers.WhoAmI)
	routeGroup.POST("/login", controllers.Login)
	routeGroup.POST("/logout", controllers.Logout)
	routeGroup.POST("/recover", controllers.RecoverUserWithInviteToken)

	routeGroup.GET("/listUsers", middlewares.RequiresAdmin(), controllers.ListUsers)
	routeGroup.POST("/createUserAndInvite", middlewares.RequiresAdmin(), controllers.CreateUserAndInvite)
	routeGroup.PUT("/:userId", middlewares.RequiresAdmin(), controllers.UpdateUser)
	routeGroup.DELETE("/:userId", middlewares.RequiresAdmin(), controllers.DeleteUser)
	routeGroup.POST("/:userId/invite", middlewares.RequiresAdmin(), controllers.CreateInviteForExistingUser)
}
