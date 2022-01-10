package api

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/up9inc/mizu/shared"
)

type ServiceMapDisabledSuite struct {
	suite.Suite

	instance ServiceMap
}

type ServiceMapEnabledSuite struct {
	suite.Suite

	instance ServiceMap
}

func (s *ServiceMapDisabledSuite) SetupTest() {
	s.instance = GetServiceMapInstance()
}

func (s *ServiceMapEnabledSuite) SetupTest() {
	s.instance = GetServiceMapInstance()
	s.instance.SetConfig(&shared.MizuAgentConfig{
		ServiceMap: true,
	})
}

func (s *ServiceMapDisabledSuite) TestServiceMapInstance() {
	assert := s.Assert()

	assert.NotNil(s.instance)
}

func (s *ServiceMapDisabledSuite) TestServiceMapSingletonInstance() {
	assert := s.Assert()

	instance2 := GetServiceMapInstance()

	assert.NotNil(s.instance)
	assert.NotNil(instance2)
	assert.Equal(s.instance, instance2)
}

func (s *ServiceMapDisabledSuite) TestServiceMapIsEnabledShouldReturnFalseByDefault() {
	assert := s.Assert()

	enabled := s.instance.IsEnabled()

	assert.False(enabled)
}

func (s *ServiceMapDisabledSuite) TestGetStatusShouldReturnDisabledByDefault() {
	assert := s.Assert()

	status := s.instance.GetStatus()

	assert.Equal("disabled", status.Status)
	assert.Equal(0, status.EntriesProcessedCount)
	assert.Equal(0, status.NodeCount)
	assert.Equal(0, status.EdgeCount)
}

func (s *ServiceMapDisabledSuite) TestAddEdgeShouldDoNothingWhenDisabled() {
	assert := s.Assert()

	s.instance.AddEdge("a", "b", "p")
	s.instance.AddEdge("c", "d", "p")
	status := s.instance.GetStatus()

	assert.Equal("disabled", status.Status)
	assert.Equal(0, status.EntriesProcessedCount)
	assert.Equal(0, status.NodeCount)
	assert.Equal(0, status.EdgeCount)
}

// Enabled

func (s *ServiceMapEnabledSuite) TestServiceMapIsEnabled() {
	assert := s.Assert()

	enabled := s.instance.IsEnabled()

	assert.True(enabled)
}

func (s *ServiceMapEnabledSuite) TestServiceMap() {
	assert := s.Assert()

	s.instance.AddEdge("a", "b", "p")
	s.instance.AddEdge("a", "b", "p")
	s.instance.AddEdge("", "a", "p")
	s.instance.AddEdge("b", "", "p")
	s.instance.AddEdge("c", "d", "p")

	status := s.instance.GetStatus()
	expectedEntriesProcessedCount := 5
	expectedNodeCount := 5
	expectedEdgeCount := 4

	// Counts
	assert.Equal(expectedEntriesProcessedCount, s.instance.GetEntriesProcessedCount())
	assert.Equal(expectedNodeCount, s.instance.GetNodesCount())
	assert.Equal(expectedEdgeCount, s.instance.GetEdgesCount())

	// Status
	assert.Equal("enabled", status.Status)
	assert.Equal(expectedEntriesProcessedCount, status.EntriesProcessedCount)
	assert.Equal(expectedNodeCount, status.NodeCount)
	assert.Equal(expectedEdgeCount, status.EdgeCount)

	// Reset
	s.instance.Reset()
	status = s.instance.GetStatus()

	// Counts after reset
	assert.Equal(0, s.instance.GetEntriesProcessedCount())
	assert.Equal(0, s.instance.GetNodesCount())
	assert.Equal(0, s.instance.GetEdgesCount())

	// Status after reset
	assert.Equal("enabled", status.Status)
	assert.Equal(0, status.EntriesProcessedCount)
	assert.Equal(0, status.NodeCount)
	assert.Equal(0, status.EdgeCount)

}

func TestServiceMapSuite(t *testing.T) {
	suite.Run(t, new(ServiceMapDisabledSuite))
	suite.Run(t, new(ServiceMapEnabledSuite))
}
