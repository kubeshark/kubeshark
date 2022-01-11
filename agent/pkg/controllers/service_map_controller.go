package controllers

import (
	"mizuserver/pkg/api"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
)

type ServiceMapController struct {
	service api.ServiceMap
}

func NewServiceMapController() *ServiceMapController {
	return &ServiceMapController{
		service: api.GetServiceMapInstance(),
	}
}

func (s *ServiceMapController) Status(c *gin.Context) {
	c.JSON(http.StatusOK, s.service.GetStatus())
}

func (s *ServiceMapController) Get(c *gin.Context) {
	response := &shared.ServiceMapResponse{
		Status: s.service.GetStatus(),
		Nodes:  s.service.GetNodes(),
		Edges:  s.service.GetEdges(),
	}
	c.JSON(http.StatusOK, response)
}

func (s *ServiceMapController) Reset(c *gin.Context) {
	s.service.Reset()
	s.Status(c)
}
