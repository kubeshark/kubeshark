package controllers

import (
	"net/http"

	core "k8s.io/api/core/v1"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/holder"
	"github.com/up9inc/mizu/agent/pkg/providers"
	"github.com/up9inc/mizu/agent/pkg/providers/tappedPods"
	"github.com/up9inc/mizu/agent/pkg/providers/tappers"
	"github.com/up9inc/mizu/agent/pkg/validation"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
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
	var requestTappedPods []core.Pod
	if err := c.Bind(&requestTappedPods); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	podInfos := kubernetes.GetPodInfosForPods(requestTappedPods)

	logger.Log.Infof("[Status] POST request: %d tapped pods", len(requestTappedPods))
	tappedPods.Set(podInfos)
	api.BroadcastTappedPodsStatus()

	nodeToTappedPodMap := kubernetes.GetNodeHostToTappedPodsMap(requestTappedPods)
	tappedPods.SetNodeToTappedPodMap(nodeToTappedPodMap)
	api.BroadcastTappedPodsToTappers(nodeToTappedPodMap)
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

func GetTappingStatus(c *gin.Context) {
	tappedPodsStatus := tappedPods.GetTappedPodsStatus()
	c.JSON(http.StatusOK, tappedPodsStatus)
}

func GetGeneralStats(c *gin.Context) {
	c.JSON(http.StatusOK, providers.GetGeneralStats())
}

func GetCurrentResolvingInformation(c *gin.Context) {
	c.JSON(http.StatusOK, holder.GetResolver().GetMap())
}
