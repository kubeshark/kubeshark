package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
	"mizuserver/pkg/up9"
	"net/http"
)

var TapStatus shared.TapStatus

func GetTappingStatus(c *gin.Context) {
	c.JSON(http.StatusOK, TapStatus)
}

func AnalyzeInformation(c *gin.Context) {
	c.JSON(http.StatusOK, up9.GetAnalyzeInfo())
}
