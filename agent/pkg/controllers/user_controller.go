package controllers

import (
	"mizuserver/pkg/providers"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared/logger"

	ory "github.com/ory/kratos-client-go"
)

func Login(c *gin.Context) {
	if token, err := providers.PerformLogin(c.PostForm("username"), c.PostForm("password"), c.Request.Context()); err != nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "bad login"})
	} else {
		c.JSON(200, gin.H{"token": token})
	}
}

func Logout(c *gin.Context) {
	token := c.GetHeader("x-session-token")
	if err := providers.Logout(token, c.Request.Context()); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "error occured while logging out, the session might still be valid"})
	} else {
		c.JSON(200, "")
	}
}

func Register(c *gin.Context) {
	token, _, err, formErrorMessages := providers.RegisterUser(c.PostForm("username"), c.PostForm("password"), c.Request.Context())
	handleRegistration(token, err, formErrorMessages, c)
}

func handleRegistration(token *string, err error, formErrorMessages map[string][]ory.UiText, c *gin.Context) {
	if err != nil {
		if formErrorMessages != nil {
			logger.Log.Infof("user attempted to register but had form errors %v %v", formErrorMessages, err)
			c.AbortWithStatusJSON(400, formErrorMessages)
		} else {
			logger.Log.Errorf("unknown internal error registering user %s", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "internal error occured while registering"})
		}
	} else {
		c.JSON(201, gin.H{"token": token})
	}
}
