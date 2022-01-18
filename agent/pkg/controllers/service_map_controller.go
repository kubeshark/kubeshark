package controllers

import (
	service "mizuserver/pkg/service_map"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceMapController struct {
	service service.ServiceMap
}

func NewServiceMapController() *ServiceMapController {
	return &ServiceMapController{
		service: service.GetServiceMapInstance(),
	}
}

func (s *ServiceMapController) Status(c *gin.Context) {
	c.JSON(http.StatusOK, s.service.GetStatus())
}

func (s *ServiceMapController) Get(c *gin.Context) {
	response := &service.ServiceMapResponse{
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
