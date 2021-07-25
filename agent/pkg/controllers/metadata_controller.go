package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/version"
	"net/http"
)

func GetVersion(c *gin.Context) {
	resp := shared.VersionResponse{SemVer: version.SemVer}
	c.JSON(http.StatusOK, resp)
}
