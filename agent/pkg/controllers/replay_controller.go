package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/validation"
	"github.com/up9inc/mizu/shared"
)

func ReplayRequest(c *gin.Context) {
	replayDetails := &shared.ReplayDetails{}
	if err := c.Bind(replayDetails); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if err := validation.Validate(replayDetails); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// TODO: execute the request with those details

	c.JSON(http.StatusOK, nil)
}
