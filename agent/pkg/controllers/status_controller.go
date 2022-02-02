package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/holder"
	"github.com/up9inc/mizu/agent/pkg/providers"
	"github.com/up9inc/mizu/agent/pkg/providers/tappedPods"
	"github.com/up9inc/mizu/agent/pkg/providers/tappers"
	"github.com/up9inc/mizu/agent/pkg/providers/user"
	"github.com/up9inc/mizu/agent/pkg/up9"
	"github.com/up9inc/mizu/agent/pkg/validation"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"

	"github.com/gin-gonic/gin"
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
	tappers.SetStatus(tapperStatus)
	broadcastTappedPodsStatus()
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
	var err error
	if tappedPodsStatus, err = filterTapStatusForUser(tappedPodsStatus, c.GetHeader(user.SessionTokenHeader), c.Request.Context()); err != nil {
		logger.Log.Errorf("Could not filter tapped pods status for user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown error occured"})
	} else {
		c.JSON(http.StatusOK, tappedPodsStatus)
	}
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

func filterTapStatusForUser(tappedPodStatus []shared.TappedPodStatus, sessionToken string, ctx context.Context) ([]shared.TappedPodStatus, error) {
	user, err := user.WhoAmI(sessionToken, ctx)
	if err != nil {
		return nil, err
	}

	if user.Workspace == nil {
		return make([]shared.TappedPodStatus, 0), nil
	}

	filteredTappedPodStatus := make([]shared.TappedPodStatus, 0)
	for _, tappedPod := range tappedPodStatus {
		if shared.Contains(user.Workspace.Namespaces, tappedPod.Namespace) {
			filteredTappedPodStatus = append(filteredTappedPodStatus, tappedPod)
		}
	}

	return filteredTappedPodStatus, nil
}
