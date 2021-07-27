package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
)

var TapStatus shared.TapStatus

func GetTappingStatus(c *gin.Context) {
	c.JSON(http.StatusOK, TapStatus)
}

func AnalyzeInformation(c *gin.Context) {
}
