package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/up9inc/mizu/agent/pkg/dependency"
	"github.com/up9inc/mizu/agent/pkg/servicemap"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	tapApi "github.com/up9inc/mizu/tap/api"
)

const (
	a    = "aService"
	b    = "bService"
	Ip   = "127.0.0.1"
	Port = "80"
)

var (
	TCPEntryA = &tapApi.TCP{
		Name: a,
		Port: Port,
		IP:   fmt.Sprintf("%s.%s", Ip, a),
	}
	TCPEntryB = &tapApi.TCP{
		Name: b,
		Port: Port,
		IP:   fmt.Sprintf("%s.%s", Ip, b),
	}
)

var ProtocolHttp = &tapApi.Protocol{
	Name:            "http",
	LongName:        "Hypertext Transfer Protocol -- HTTP/1.1",
	Abbreviation:    "HTTP",
	Macro:           "http",
	Version:         "1.1",
	BackgroundColor: "#205cf5",
	ForegroundColor: "#ffffff",
	FontSize:        12,
	ReferenceLink:   "https://datatracker.ietf.org/doc/html/rfc2616",
	Ports:           []string{"80", "443", "8080"},
	Priority:        0,
}

type ServiceMapControllerSuite struct {
	suite.Suite

	c *ServiceMapController
	w *httptest.ResponseRecorder
	g *gin.Context
}

func (s *ServiceMapControllerSuite) SetupTest() {
	dependency.RegisterGenerator(dependency.ServiceMapGeneratorDependency, func() interface{} { return servicemap.GetDefaultServiceMapInstance() })

	s.c = NewServiceMapController()
	s.c.service.Enable()
	s.c.service.(servicemap.ServiceMapSink).NewTCPEntry(TCPEntryA, TCPEntryB, ProtocolHttp)

	s.w = httptest.NewRecorder()
	s.g, _ = gin.CreateTestContext(s.w)
}

func (s *ServiceMapControllerSuite) TestGetStatus() {
	assert := s.Assert()

	s.c.Status(s.g)
	assert.Equal(http.StatusOK, s.w.Code)

	var status servicemap.ServiceMapStatus
	err := json.Unmarshal(s.w.Body.Bytes(), &status)
	assert.NoError(err)
	assert.Equal("enabled", status.Status)
	assert.Equal(1, status.EntriesProcessedCount)
	assert.Equal(2, status.NodeCount)
	assert.Equal(1, status.EdgeCount)
}

func (s *ServiceMapControllerSuite) TestGet() {
	assert := s.Assert()

	s.c.Get(s.g)
	assert.Equal(http.StatusOK, s.w.Code)

	var response servicemap.ServiceMapResponse
	err := json.Unmarshal(s.w.Body.Bytes(), &response)
	assert.NoError(err)

	// response status
	assert.Equal("enabled", response.Status.Status)
	assert.Equal(1, response.Status.EntriesProcessedCount)
	assert.Equal(2, response.Status.NodeCount)
	assert.Equal(1, response.Status.EdgeCount)

	// response nodes
	aNode := servicemap.ServiceMapNode{
		Id:       1,
		Name:     TCPEntryA.Name,
		Entry:    TCPEntryA,
		Resolved: true,
		Count:    1,
	}
	bNode := servicemap.ServiceMapNode{
		Id:       2,
		Name:     TCPEntryB.Name,
		Entry:    TCPEntryB,
		Resolved: true,
		Count:    1,
	}
	assert.Contains(response.Nodes, aNode)
	assert.Contains(response.Nodes, bNode)
	assert.Len(response.Nodes, 2)

	// response edges
	assert.Equal([]servicemap.ServiceMapEdge{
		{
			Source:      aNode,
			Destination: bNode,
			Protocol:    ProtocolHttp,
			Count:       1,
		},
	}, response.Edges)
}

func (s *ServiceMapControllerSuite) TestGetReset() {
	assert := s.Assert()

	s.c.Reset(s.g)
	assert.Equal(http.StatusOK, s.w.Code)

	var status servicemap.ServiceMapStatus
	err := json.Unmarshal(s.w.Body.Bytes(), &status)
	assert.NoError(err)
	assert.Equal("enabled", status.Status)
	assert.Equal(0, status.EntriesProcessedCount)
	assert.Equal(0, status.NodeCount)
	assert.Equal(0, status.EdgeCount)
}

func TestServiceMapControllerSuite(t *testing.T) {
	suite.Run(t, new(ServiceMapControllerSuite))
}
