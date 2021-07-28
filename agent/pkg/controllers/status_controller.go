package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/tap"
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
	recentTLSLinks := make([]string, 0)

	for _, outboundLinkItem := range providers.RecentTLSLinks.Items() {
		outboundLink, castOk := outboundLinkItem.Object.(*tap.OutboundLink)
		if castOk {
			recentTLSLinks = append(recentTLSLinks, outboundLink.DstIP)
		}
	}

	c.JSON(http.StatusOK, recentTLSLinks)
}
