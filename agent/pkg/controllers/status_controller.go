package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/api"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/validation"
	"net/http"
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
