package controllers

import (
	"mizuserver/pkg/api"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
)

type ServiceMapController struct{}

func NewServiceMapController() *ServiceMapController {
	return &ServiceMapController{}
}

func (s *ServiceMapController) Status(c *gin.Context) {
	serviceMap := api.GetServiceMapInstance()
	c.JSON(http.StatusOK, serviceMap.GetStatus())
}

func (s *ServiceMapController) Get(c *gin.Context) {
	serviceMap := api.GetServiceMapInstance()
	response := &shared.ServiceMapResponse{
		Status: serviceMap.GetStatus(),
		Nodes:  serviceMap.GetNodes(),
		Edges:  serviceMap.GetEdges(),
	}
	c.JSON(http.StatusOK, response)
}

func (s *ServiceMapController) Reset(c *gin.Context) {
	serviceMap := api.GetServiceMapInstance()
	serviceMap.Reset()
	s.Status(c)
}
