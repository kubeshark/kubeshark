package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// NotFoundRoute defines the 404 Error route.
func NotFoundRoute(app *gin.Engine) {
	app.Use(
		func(c *gin.Context) {
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": true,
				"msg":   "sorry, endpoint is not found",
			})
		},
	)
}
