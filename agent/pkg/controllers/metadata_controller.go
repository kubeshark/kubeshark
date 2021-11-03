package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/config"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/version"
	"net/http"
)

func GetVersion(c *gin.Context) {
	resp := shared.VersionResponse{SemVer: version.SemVer}
	c.JSON(http.StatusOK, resp)
}

func HealthCheck(c *gin.Context) {
	if config.Config.DaemonMode && (providers.TapStatus.Pods == nil || len(providers.TapStatus.Pods) == 0) {
		c.String(http.StatusInternalServerError, "no tapped pods")
	}
	if providers.TappersCount == 0 {
		c.String(http.StatusInternalServerError, "no tappers are connected")
	}
	response := shared.HealthResponse{
		TapStatus:    providers.TapStatus,
		TappersCount: providers.TappersCount,
	}
	c.JSON(http.StatusOK, response)
}
