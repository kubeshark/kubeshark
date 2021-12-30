package controllers

import (
	"mizuserver/pkg/providers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsSetupNecessary(c *gin.Context) {
	if IsInstallNeeded, err := providers.IsInstallNeeded(); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "internal error occured while checking if install is needed"})
	} else {
		c.JSON(http.StatusOK, IsInstallNeeded)
	}
}

func Setup(c *gin.Context) {
	if err := providers.DoInstall(c.PostForm("admin_password"), c.Request.Context()); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "internal error occured while setting up"})
	} else {
		c.JSON(http.StatusCreated, "")
	}
}
