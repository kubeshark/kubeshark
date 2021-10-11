package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/api"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/validation"
	"net/http"
	"os"
)

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
	rlog.Infof("[Status] POST request: %d tapped pods", len(tapStatus.Pods))
	providers.TapStatus.Pods = tapStatus.Pods
	message := shared.CreateWebSocketStatusMessage(*tapStatus)
	if jsonBytes, err := json.Marshal(message); err != nil {
		rlog.Errorf("Could not Marshal message %v\n", err)
	} else {
		api.BroadcastToBrowserClients(jsonBytes)
	}
}

func GetTappersCount(c *gin.Context) {
	c.JSON(http.StatusOK, providers.TappersCount)
}

func GetAuthStatus(c *gin.Context) {
	authStatusJson := os.Getenv(shared.AuthStatusEnvVar)
	if authStatusJson == "" {
		authStatus := shared.AuthStatus{}
		c.JSON(http.StatusOK, authStatus)
		return
	}

	var authStatus shared.AuthStatus
	err := json.Unmarshal([]byte(authStatusJson), &authStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal auth status, err: %v", err))
		return
	}

	c.JSON(http.StatusOK, authStatus)
}
