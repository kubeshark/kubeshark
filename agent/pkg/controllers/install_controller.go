package controllers

import (
	"mizuserver/pkg/providers"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared/logger"
)

func IsSetupNecessary(c *gin.Context) {
	if IsInstallNeeded, err := providers.IsInstallNeeded(); err != nil {
		logger.Log.Errorf("unknown internal while checking if install is needed %s", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal error occured while checking if install is needed"})
	} else {
		c.JSON(http.StatusOK, IsInstallNeeded)
	}
}
