package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/replay"
	"github.com/up9inc/mizu/agent/pkg/validation"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
)

func ReplayRequest(c *gin.Context) {
	fmt.Print("Starting replay")
	replayDetails := &shared.ReplayDetails{}
	if err := c.Bind(replayDetails); err != nil {
		logger.Log.Errorf("ERR1 %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	logger.Log.Warningf("Validating replay, %+v", replayDetails)
	if err := validation.Validate(replayDetails); err != nil {
		logger.Log.Errorf("ERR2 %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	logger.Log.Warningf("Executing replay")
	resp, err := replay.ExecuteRequest(replayDetails)
	if err != nil {
		logger.Log.Errorf("ERR3 %v", err)
		c.JSON(http.StatusBadRequest, err.Error())
	}

	logger.Log.Infof("Result %v", resp)
	c.JSON(http.StatusOK, resp)
}
