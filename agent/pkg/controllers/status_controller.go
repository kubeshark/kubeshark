package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	"mizuserver/pkg/api"
	"mizuserver/pkg/holder"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/providers/tappedPods"
	"mizuserver/pkg/providers/tappersCount"
	"mizuserver/pkg/providers/tappersStatus"
	"mizuserver/pkg/up9"
	"mizuserver/pkg/validation"
	"net/http"
)

func HealthCheck(c *gin.Context) {
	tappers := make([]*shared.TapperStatus, 0)
	for _, value := range tappersStatus.Get() {
		tappers = append(tappers, value)
	}

	response := shared.HealthResponse{
		TappedPods:    tappedPods.Get(),
		TappersCount:  tappersCount.Get(),
		TappersStatus: tappers,
	}
	c.JSON(http.StatusOK, response)
}

func PostTappedPods(c *gin.Context) {
	var requestTappedPods []*shared.PodInfo
	if err := c.Bind(&requestTappedPods); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	logger.Log.Infof("[Status] POST request: %d tapped pods", len(requestTappedPods))
	tappedPods.Set(requestTappedPods)
	broadcastTappedPodsStatus()
}

func broadcastTappedPodsStatus() {
	tappedPodsStatus := tappedPods.GetTappedPodsStatus()

	message := shared.CreateWebSocketStatusMessage(tappedPodsStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		logger.Log.Errorf("Could not Marshal message %v", err)
	} else {
		api.BroadcastToBrowserClients(jsonBytes)
	}
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
	tappersStatus.Set(tapperStatus)
	broadcastTappedPodsStatus()
}

func GetTappersCount(c *gin.Context) {
	c.JSON(http.StatusOK, tappersCount.Get())
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
	tappedPodsStatus := tappedPods.GetTappedPodsStatus()
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
