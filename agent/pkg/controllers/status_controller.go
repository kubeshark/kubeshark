package controllers

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/providers"
	"mizuserver/pkg/up9"
	"net/http"
)

func GetTappingStatus(c *gin.Context) {
	c.JSON(http.StatusOK, providers.TapStatus)
}

func AnalyzeInformation(c *gin.Context) {
	c.JSON(http.StatusOK, up9.GetAnalyzeInfo())
}

func GetRecentTLSLinks(c *gin.Context) {
	c.JSON(http.StatusOK, providers.GetAllRecentTLSAddresses())
}
