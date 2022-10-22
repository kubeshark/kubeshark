package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/kubeshark/agent/pkg/version"
	"github.com/up9inc/kubeshark/shared"
)

func GetVersion(c *gin.Context) {
	resp := shared.VersionResponse{Ver: version.Ver}
	c.JSON(http.StatusOK, resp)
}
