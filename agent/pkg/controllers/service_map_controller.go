package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceMapController struct{}

func NewServiceMapController() *ServiceMapController {
	return &ServiceMapController{}
}

func (s *ServiceMapController) Status(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

func (s *ServiceMapController) Get(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}

func (s *ServiceMapController) Reset(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, nil)
}
