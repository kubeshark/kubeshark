package routes

import (
	"mizuserver/pkg/controllers"
	"mizuserver/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

func UserRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/user")

	routeGroup.GET("/:userId", middlewares.RequiresAdmin(), controllers.GetUser)

	routeGroup.POST("/login", controllers.Login)
	routeGroup.POST("/logout", controllers.Logout)
	routeGroup.POST("/register", controllers.RegisterWithToken)

	routeGroup.GET("/listUsers", middlewares.RequiresAdmin(), controllers.ListUsers)
	routeGroup.POST("/createUserAndInvite", middlewares.RequiresAdmin(), controllers.CreateUserAndInvite)
	routeGroup.PUT("/:userId", middlewares.RequiresAdmin(), controllers.UpdateUser)
	routeGroup.DELETE("/:userId", middlewares.RequiresAdmin(), controllers.DeleteUser)
}
