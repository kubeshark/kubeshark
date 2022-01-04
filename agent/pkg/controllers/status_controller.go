package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	"mizuserver/pkg/api"
	"mizuserver/pkg/holder"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/up9"
	"mizuserver/pkg/utils"
	"mizuserver/pkg/validation"
	"net/http"
)

func HealthCheck(c *gin.Context) {
	tappers := make([]shared.TapperStatus, 0)
	for _, value := range providers.TappersStatus {
		tappers = append(tappers, value)
	}

	response := shared.HealthResponse{
		TapStatus:     providers.TapStatus,
		TappersCount:  providers.TappersCount,
		TappersStatus: tappers,
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
	broadcastTappedPodsStatus()
}

func broadcastTappedPodsStatus() {
	tappedPodsStatus := utils.GetTappedPodsStatus()

	message := shared.CreateWebSocketStatusMessage(tappedPodsStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		logger.Log.Errorf("Could not Marshal message %v", err)
	} else {
		api.BroadcastToBrowserClients(jsonBytes)
	}
}

func addTapperStatus(tapperStatus shared.TapperStatus) {
	if providers.TappersStatus == nil {
		providers.TappersStatus = make(map[string]shared.TapperStatus)
	}

	providers.TappersStatus[tapperStatus.NodeName] = tapperStatus
}

func PostTapperStatus(c *gin.Context) {
	tapperStatus := &shared.TapperStatus{}
	if err := c.Bind(tapperStatus); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if err := validation.Validate(tapperStatus); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	logger.Log.Infof("[Status] POST request, tapper status: %v", tapperStatus)
	addTapperStatus(*tapperStatus)
	broadcastTappedPodsStatus()
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
	tappedPodsStatus := utils.GetTappedPodsStatus()
	c.JSON(http.StatusOK, tappedPodsStatus)
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
