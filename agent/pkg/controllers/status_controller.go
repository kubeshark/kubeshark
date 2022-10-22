package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	core "k8s.io/api/core/v1"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/kubeshark/agent/pkg/api"
	"github.com/kubeshark/kubeshark/agent/pkg/holder"
	"github.com/kubeshark/kubeshark/agent/pkg/providers"
	"github.com/kubeshark/kubeshark/agent/pkg/providers/tappedPods"
	"github.com/kubeshark/kubeshark/agent/pkg/providers/tappers"
	"github.com/kubeshark/kubeshark/agent/pkg/validation"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared"
	"github.com/kubeshark/kubeshark/shared/kubernetes"
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

func GetTrafficStats(c *gin.Context) {
	startTime, endTime, err := getStartEndTime(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, providers.GetTrafficStats(startTime, endTime))
}

func getStartEndTime(c *gin.Context) (time.Time, time.Time, error) {
	startTimeValue, err := strconv.Atoi(c.Query("startTimeMs"))
	if err != nil {
		return time.UnixMilli(0), time.UnixMilli(0), fmt.Errorf("invalid start time: %v", err)
	}
	endTimeValue, err := strconv.Atoi(c.Query("endTimeMs"))
	if err != nil {
		return time.UnixMilli(0), time.UnixMilli(0), fmt.Errorf("invalid end time: %v", err)
	}
	return time.UnixMilli(int64(startTimeValue)), time.UnixMilli(int64(endTimeValue)), nil
}

func GetCurrentResolvingInformation(c *gin.Context) {
	c.JSON(http.StatusOK, holder.GetResolver().GetMap())
}
