package controllers

import (
	"mizuserver/pkg/version"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
)

func GetVersion(c *gin.Context) {
	resp := shared.VersionResponse{SemVer: version.SemVer}
	c.JSON(http.StatusOK, resp)
}
