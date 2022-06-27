package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/replay"
	"github.com/up9inc/mizu/agent/pkg/validation"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
)

func ReplayRequest(c *gin.Context) {
	logger.Log.Debug("Starting replay")
	replayDetails := &shared.ReplayDetails{}
	if err := c.Bind(replayDetails); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	logger.Log.Debugf("Validating replay, %v", replayDetails)
	if err := validation.Validate(replayDetails); err != nil {
		logger.Log.Errorf("Error Validating replay details object %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	logger.Log.Debug("Executing replay")
	resultChannel := make(chan *shared.ReplayResponse, 1)
	replay.ExecuteRequest(replayDetails, resultChannel)
	result := <-resultChannel
	c.JSON(http.StatusOK, result)
}
