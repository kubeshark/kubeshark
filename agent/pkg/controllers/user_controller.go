package controllers

import (
	"errors"
	"mizuserver/pkg/providers"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared/logger"
)

func Login(c *gin.Context) {
	if token, err := providers.PerformLogin(c.PostForm("username"), c.PostForm("password"), c.Request.Context()); err != nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "bad login"})
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("x-session-token", *token, 3600, "/", "", false, false)
		c.JSON(200, gin.H{"token": token})
	}
}

func Logout(c *gin.Context) {
	token, err := c.Cookie("x-session-token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			c.AbortWithStatusJSON(401, gin.H{"error": "could not find session cookie"})
		} else {
			logger.Log.Errorf("error reading cookie in logout %s", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "error occured while logging out, the session might still be valid"})
		}

		return
	}

	if err = providers.Logout(token, c.Request.Context()); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "error occured while logging out, the session might still be valid"})
	} else {
		c.SetCookie("x-session-token", "", -1, "/", "", false, false)
		c.JSON(200, "")
	}
}

func Register(c *gin.Context) {
	// only allow one user to be created without authentication
	if IsInstallNeeded, err := providers.IsInstallNeeded(); err != nil {
		logger.Log.Errorf("unknown internal while checking if install is needed %s", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal error occured while checking if install is needed"})
		return
	} else if !IsInstallNeeded {
		c.AbortWithStatusJSON(401, gin.H{"error": "cannot register when install is not needed"})
		return
	}

	if token, _, err, formErrorMessages := providers.RegisterUser(c.PostForm("username"), c.PostForm("password"), c.Request.Context()); err != nil {
		if formErrorMessages != nil {
			logger.Log.Infof("user attempted to register but had form errors %v %v", formErrorMessages, err)
			c.AbortWithStatusJSON(400, formErrorMessages)
		} else {
			logger.Log.Errorf("unknown internal error registering user %s", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "internal error occured while registering"})
		}
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("x-session-token", *token, 3600, "/", "", false, false)
		c.JSON(200, gin.H{"token": token})
	}
}
