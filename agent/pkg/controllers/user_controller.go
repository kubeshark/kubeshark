package controllers

import (
	"mizuserver/pkg/providers"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	if token, err := providers.PerformLogin(c.PostForm("username"), c.PostForm("password"), c.Request.Context()); err != nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "bad login"})
	} else {
		c.JSON(200, gin.H{"token": token})
	}
}
