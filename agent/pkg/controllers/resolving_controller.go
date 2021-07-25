package controllers

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/holder"
	"net/http"
)

func GetCurrentResolvingInformation(c *gin.Context) {
	c.JSON(http.StatusOK, holder.GetResolver().GetMap())
}

