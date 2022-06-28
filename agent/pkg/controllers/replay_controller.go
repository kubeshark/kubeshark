package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/replay"
	"github.com/up9inc/mizu/agent/pkg/validation"
	"github.com/up9inc/mizu/logger"
)

const (
	replayTimeout = 10 * time.Second
)

func ReplayRequest(c *gin.Context) {
	logger.Log.Debug("Starting replay")
	replayDetails := &replay.Details{}
	if err := c.Bind(replayDetails); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	logger.Log.Debugf("Validating replay, %v", replayDetails)
	if err := validation.Validate(replayDetails); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	logger.Log.Debug("Executing replay, %v", replayDetails)
	result := replay.ExecuteRequest(replayDetails, replayTimeout)
	c.JSON(http.StatusOK, result)
}
