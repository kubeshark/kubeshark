package routes

import (
	"mizuserver/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/user")

	routeGroup.POST("/login", controllers.Login)
	routeGroup.POST("/logout", controllers.Logout)
	routeGroup.POST("/register", controllers.Register)
}
