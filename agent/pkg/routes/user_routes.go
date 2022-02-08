package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"

	"github.com/gin-gonic/gin"
)

var (
	UserPostLogin    = controllers.Login
	UserPostLogout   = controllers.Logout
	UserPostRegister = controllers.Register
)

func UserRoutes(ginApp *gin.Engine) *gin.RouterGroup {
	routeGroup := ginApp.Group("/user")

	routeGroup.POST("/login", func(c *gin.Context) { UserPostLogin(c) })
	routeGroup.POST("/logout", func(c *gin.Context) { UserPostLogout(c) })
	routeGroup.POST("/register", func(c *gin.Context) { UserPostRegister(c) })

	return routeGroup
}
