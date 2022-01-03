package controllers

import (
	"mizuserver/pkg/providers"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared/logger"
)

func IsSetupNecessary(c *gin.Context) {
	if IsInstallNeeded, err := providers.IsInstallNeeded(); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "internal error occured while checking if install is needed"})
	} else {
		c.JSON(http.StatusOK, IsInstallNeeded)
	}
}

func Setup(c *gin.Context) {
	if token, err := providers.DoInstall(c.PostForm("adminPassword"), c.Request.Context()); err != nil {
		logger.Log.Error(err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal error occured while setting up"})
	} else {
		c.JSON(http.StatusCreated, gin.H{"token": token})
	}
}
