package controllers

import (
	"net/http"

	"github.com/up9inc/mizu/agent/pkg/dependency"
	"github.com/up9inc/mizu/agent/pkg/servicemap"

	"github.com/gin-gonic/gin"
)

type ServiceMapController struct {
	service servicemap.ServiceMap
}

func NewServiceMapController() *ServiceMapController {
	serviceMapGenerator := dependency.GetInstance(dependency.ServiceMapGeneratorDependency).(servicemap.ServiceMap)
	return &ServiceMapController{
		service: serviceMapGenerator,
	}
}

func (s *ServiceMapController) Status(c *gin.Context) {
	c.JSON(http.StatusOK, s.service.GetStatus())
}

func (s *ServiceMapController) Get(c *gin.Context) {
	response := &servicemap.ServiceMapResponse{
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
