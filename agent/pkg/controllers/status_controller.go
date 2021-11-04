package controllers

import (
	"encoding/json"
	"mizuserver/pkg/api"
	"mizuserver/pkg/holder"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/up9"
	"mizuserver/pkg/validation"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

func HealthCheck(c *gin.Context) {
	response := shared.HealthResponse{
		TapStatus:    providers.TapStatus,
		TappersCount: providers.TappersCount,
	}
	c.JSON(http.StatusOK, response)
}


func PostTappedPods(c *gin.Context) {
	tapStatus := &shared.TapStatus{}
	if err := c.Bind(tapStatus); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if err := validation.Validate(tapStatus); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	logger.Log.Infof("[Status] POST request: %d tapped pods", len(tapStatus.Pods))
	providers.TapStatus.Pods = tapStatus.Pods
	message := shared.CreateWebSocketStatusMessage(*tapStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		logger.Log.Errorf("Could not Marshal message %v\n", err)
	} else {
		api.BroadcastToBrowserClients(jsonBytes)
	}
}

func GetTappersCount(c *gin.Context) {
	c.JSON(http.StatusOK, providers.TappersCount)
}

func GetAuthStatus(c *gin.Context) {
	authStatus, err := providers.GetAuthStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, authStatus)
}

func GetTappingStatus(c *gin.Context) {
	c.JSON(http.StatusOK, providers.TapStatus)
}

func AnalyzeInformation(c *gin.Context) {
	c.JSON(http.StatusOK, up9.GetAnalyzeInfo())
}

func GetGeneralStats(c *gin.Context) {
	c.JSON(http.StatusOK, providers.GetGeneralStats())
}

func GetRecentTLSLinks(c *gin.Context) {
	c.JSON(http.StatusOK, providers.GetAllRecentTLSAddresses())
}

func GetCurrentResolvingInformation(c *gin.Context) {
	c.JSON(http.StatusOK, holder.GetResolver().GetMap())
}
