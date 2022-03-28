package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/holder"
	"github.com/up9inc/mizu/agent/pkg/providers"
	"github.com/up9inc/mizu/agent/pkg/providers/tappedPods"
	"github.com/up9inc/mizu/agent/pkg/providers/tappers"
	"github.com/up9inc/mizu/agent/pkg/up9"
	"github.com/up9inc/mizu/agent/pkg/validation"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

func HealthCheck(c *gin.Context) {
	tappersStatus := make([]*shared.TapperStatus, 0)
	for _, value := range tappers.GetStatus() {
		tappersStatus = append(tappersStatus, value)
	}

	response := shared.HealthResponse{
		TappedPods:            tappedPods.Get(),
		ConnectedTappersCount: tappers.GetConnectedCount(),
		TappersStatus:         tappersStatus,
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
	api.BroadcastTappedPodsStatus()
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
	tappers.SetStatus(tapperStatus)
	api.BroadcastTappedPodsStatus()
}

func GetConnectedTappersCount(c *gin.Context) {
	c.JSON(http.StatusOK, tappers.GetConnectedCount())
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
