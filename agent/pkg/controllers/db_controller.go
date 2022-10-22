package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/kubeshark/agent/pkg/app"
	"github.com/kubeshark/kubeshark/agent/pkg/config"
	"github.com/kubeshark/kubeshark/shared"
	basenine "github.com/up9inc/basenine/client/go"
)

func Flush(c *gin.Context) {
	if err := basenine.Flush(shared.BasenineHost, shared.BaseninePort); err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, "Flushed.")
	}
}

func Reset(c *gin.Context) {
	if err := basenine.Reset(shared.BasenineHost, shared.BaseninePort); err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		app.ConfigureBasenineServer(shared.BasenineHost, shared.BaseninePort, config.Config.MaxDBSizeBytes, config.Config.LogLevel, config.Config.InsertionFilter)
		c.JSON(http.StatusOK, "Resetted.")
	}
}
