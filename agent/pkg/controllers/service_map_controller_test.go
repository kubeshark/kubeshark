package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/up9inc/mizu/shared"
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
	s.c = NewServiceMapController()
	s.c.service.SetConfig(&shared.MizuAgentConfig{
		ServiceMap: true,
	})
	s.c.service.NewTCPEntry(TCPEntryA, TCPEntryB, ProtocolHttp)

	s.w = httptest.NewRecorder()
	s.g, _ = gin.CreateTestContext(s.w)
}

func (s *ServiceMapControllerSuite) TestGetStatus() {
	assert := s.Assert()

	s.c.Status(s.g)
	assert.Equal(http.StatusOK, s.w.Code)

	var status shared.ServiceMapStatus
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

	var response shared.ServiceMapResponse
	err := json.Unmarshal(s.w.Body.Bytes(), &response)
	assert.NoError(err)

	// response status
	assert.Equal("enabled", response.Status.Status)
	assert.Equal(1, response.Status.EntriesProcessedCount)
	assert.Equal(2, response.Status.NodeCount)
	assert.Equal(1, response.Status.EdgeCount)

	// response nodes
	aNode := shared.ServiceMapNode{
		Id:    1,
		Name:  TCPEntryA.IP,
		Entry: TCPEntryA,
		Count: 1,
	}
	bNode := shared.ServiceMapNode{
		Id:    2,
		Name:  TCPEntryB.IP,
		Entry: TCPEntryB,
		Count: 1,
	}
	assert.Contains(response.Nodes, aNode)
	assert.Contains(response.Nodes, bNode)
	assert.Len(response.Nodes, 2)

	// response edges
	assert.Equal([]shared.ServiceMapEdge{
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

	var status shared.ServiceMapStatus
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
