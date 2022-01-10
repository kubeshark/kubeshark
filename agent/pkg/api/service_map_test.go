package api

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/up9inc/mizu/shared"
)

const Protocol = "p"

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

	s.instance.AddEdge("a", "b", Protocol)
	s.instance.AddEdge("c", "d", Protocol)
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

	// 6 entries
	s.instance.AddEdge("a", "b", Protocol)
	s.instance.AddEdge("a", "b", Protocol)
	s.instance.AddEdge("", "a", Protocol)
	s.instance.AddEdge("b", "", Protocol)
	s.instance.AddEdge("c", "d", Protocol)
	s.instance.AddEdge("a", "c", Protocol)

	status := s.instance.GetStatus()
	nodes := s.instance.GetNodes()
	edges := s.instance.GetEdges()
	expectedEntriesProcessedCount := 6
	expectedNodeCount := 5
	expectedEdgeCount := 5

	// Counts
	assert.Equal(expectedEntriesProcessedCount, s.instance.GetEntriesProcessedCount())
	assert.Equal(expectedNodeCount, s.instance.GetNodesCount())
	assert.Equal(expectedEdgeCount, s.instance.GetEdgesCount())

	// Status
	assert.Equal("enabled", status.Status)
	assert.Equal(expectedEntriesProcessedCount, status.EntriesProcessedCount)
	assert.Equal(expectedNodeCount, status.NodeCount)
	assert.Equal(expectedEdgeCount, status.EdgeCount)

	// Nodes
	aNode := -1
	bNode := -1
	cNode := -1
	dNode := -1
	unresolvedNode := -1
	var validateNode = func(node shared.ServiceMapNode, index int, count int) int {
		// id
		assert.GreaterOrEqual(node.Id, 1)
		assert.LessOrEqual(node.Id, expectedNodeCount)

		// protocol
		assert.Equal(Protocol, node.Protocol)

		// count
		assert.Equal(count, node.Count)

		return node.Id
	}

	for i, v := range nodes {
		if v.Name == "a" {
			aNode = validateNode(v, i, 4)
			continue
		}
		if v.Name == "b" {
			bNode = validateNode(v, i, 3)
			continue
		}
		if v.Name == "c" {
			cNode = validateNode(v, i, 2)
			continue
		}
		if v.Name == "d" {
			dNode = validateNode(v, i, 1)
			continue
		}
		if v.Name == UnresolvedNode {
			unresolvedNode = validateNode(v, i, 2)
			continue
		}
	}

	// Make sure we found all the nodes
	nodeIds := [...]int{aNode, bNode, cNode, dNode, unresolvedNode}
	for _, v := range nodeIds {
		assert.NotEqual(-1, v)
	}

	// Edges
	abEdge := -1
	uaEdge := -1
	buEdge := -1
	cdEdge := -1
	acEdge := -1
	var validateEdge = func(edge shared.ServiceMapEdge, count int) {
		// source
		assert.Contains(nodeIds, edge.Source.Id)
		assert.LessOrEqual(edge.Source.Id, expectedNodeCount)

		// destination
		assert.Contains(nodeIds, edge.Destination.Id)
		assert.LessOrEqual(edge.Destination.Id, expectedNodeCount)

		// protocol
		assert.Equal(Protocol, edge.Source.Protocol)
		assert.Equal(Protocol, edge.Destination.Protocol)

		// count
		assert.Equal(count, edge.Count)
	}

	for i, v := range edges {
		if v.Source.Name == "a" && v.Destination.Name == "b" {
			validateEdge(v, 2)
			abEdge = i
			continue
		}
		if v.Source.Name == UnresolvedNode && v.Destination.Name == "a" {
			validateEdge(v, 1)
			uaEdge = i
			continue
		}
		if v.Source.Name == "b" && v.Destination.Name == UnresolvedNode {
			validateEdge(v, 1)
			buEdge = i
			continue
		}
		if v.Source.Name == "c" && v.Destination.Name == "d" {
			validateEdge(v, 1)
			cdEdge = i
			continue
		}
		if v.Source.Name == "a" && v.Destination.Name == "c" {
			validateEdge(v, 1)
			acEdge = i
			continue
		}
	}

	// Make sure we found all the edges
	for _, v := range [...]int{abEdge, uaEdge, buEdge, cdEdge, acEdge} {
		assert.NotEqual(-1, v)
	}

	// Reset
	s.instance.Reset()
	status = s.instance.GetStatus()
	nodes = s.instance.GetNodes()
	edges = s.instance.GetEdges()

	// Counts after reset
	assert.Equal(0, s.instance.GetEntriesProcessedCount())
	assert.Equal(0, s.instance.GetNodesCount())
	assert.Equal(0, s.instance.GetEdgesCount())

	// Status after reset
	assert.Equal("enabled", status.Status)
	assert.Equal(0, status.EntriesProcessedCount)
	assert.Equal(0, status.NodeCount)
	assert.Equal(0, status.EdgeCount)

	// Nodes after reset
	assert.Equal([]shared.ServiceMapNode(nil), nodes)

	// Edges after reset
	assert.Equal([]shared.ServiceMapEdge(nil), edges)
}

func TestServiceMapSuite(t *testing.T) {
	suite.Run(t, new(ServiceMapDisabledSuite))
	suite.Run(t, new(ServiceMapEnabledSuite))
}
