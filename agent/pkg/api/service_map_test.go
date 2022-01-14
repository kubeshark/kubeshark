package api

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/up9inc/mizu/shared"
	tapApi "github.com/up9inc/mizu/tap/api"
)

const (
	a    = "aService"
	b    = "bService"
	c    = "cService"
	d    = "dService"
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
	TCPEntryC = &tapApi.TCP{
		Name: c,
		Port: Port,
		IP:   fmt.Sprintf("%s.%s", Ip, c),
	}
	TCPEntryD = &tapApi.TCP{
		Name: d,
		Port: Port,
		IP:   fmt.Sprintf("%s.%s", Ip, d),
	}
	TCPEntryUnresolved = &tapApi.TCP{
		Name: "",
		Port: Port,
		IP:   Ip,
	}
	TCPEntryUnresolved2 = &tapApi.TCP{
		Name: "",
		Port: Port,
		IP:   fmt.Sprintf("%s.%s", Ip, UnresolvedNodeName),
	}
	Protocol = &tapApi.Protocol{
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

func (s *ServiceMapDisabledSuite) TestNewTCPEntryShouldDoNothingWhenDisabled() {
	assert := s.Assert()

	s.instance.NewTCPEntry(TCPEntryA, TCPEntryB, Protocol)
	s.instance.NewTCPEntry(TCPEntryC, TCPEntryD, Protocol)
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

	s.instance.NewTCPEntry(TCPEntryA, TCPEntryB, Protocol)
	s.instance.NewTCPEntry(TCPEntryA, TCPEntryB, Protocol)
	s.instance.NewTCPEntry(TCPEntryUnresolved, TCPEntryA, Protocol)
	s.instance.NewTCPEntry(TCPEntryB, TCPEntryUnresolved2, Protocol)
	s.instance.NewTCPEntry(TCPEntryC, TCPEntryD, Protocol)
	s.instance.NewTCPEntry(TCPEntryA, TCPEntryC, Protocol)

	status := s.instance.GetStatus()
	nodes := s.instance.GetNodes()
	edges := s.instance.GetEdges()
	expectedEntriesProcessedCount := 6
	expectedNodeCount := 6
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
	unresolvedNode2 := -1
	var validateNode = func(node shared.ServiceMapNode, entryName string, count int) int {
		// id
		assert.GreaterOrEqual(node.Id, 1)
		assert.LessOrEqual(node.Id, expectedNodeCount)

		// entry
		// node.Name is the key of the node, key = entry.IP
		// entry.Name is the name of the service and could be unresolved
		assert.Equal(node.Name, node.Entry.IP)
		assert.Equal(Port, node.Entry.Port)
		assert.Equal(entryName, node.Entry.Name)

		// protocol
		assert.Equal(Protocol, node.Protocol)

		// count
		assert.Equal(count, node.Count)

		return node.Id
	}

	for _, v := range nodes {

		if strings.HasSuffix(v.Name, a) {
			aNode = validateNode(v, a, 4)
			continue
		}
		if strings.HasSuffix(v.Name, b) {
			bNode = validateNode(v, b, 3)
			continue
		}
		if strings.HasSuffix(v.Name, c) {
			cNode = validateNode(v, c, 2)
			continue
		}
		if strings.HasSuffix(v.Name, d) {
			dNode = validateNode(v, d, 1)
			continue
		}
		if v.Name == Ip {
			unresolvedNode = validateNode(v, UnresolvedNodeName, 1)
			continue
		}
		if strings.HasSuffix(v.Name, UnresolvedNodeName) {
			unresolvedNode2 = validateNode(v, UnresolvedNodeName, 1)
			continue
		}
	}

	// Make sure we found all the nodes
	nodeIds := [...]int{aNode, bNode, cNode, dNode, unresolvedNode, unresolvedNode2}
	for _, v := range nodeIds {
		assert.NotEqual(-1, v)
	}

	// Edges
	abEdge := -1
	uaEdge := -1
	buEdge := -1
	cdEdge := -1
	acEdge := -1
	var validateEdge = func(edge shared.ServiceMapEdge, sourceEntryName string, destEntryName string, count int) {
		// source
		assert.Contains(nodeIds, edge.Source.Id)
		assert.LessOrEqual(edge.Source.Id, expectedNodeCount)
		assert.Equal(edge.Source.Name, edge.Source.Entry.IP)
		assert.Equal(sourceEntryName, edge.Source.Entry.Name)

		// destination
		assert.Contains(nodeIds, edge.Destination.Id)
		assert.LessOrEqual(edge.Destination.Id, expectedNodeCount)
		assert.Equal(edge.Destination.Name, edge.Destination.Entry.IP)
		assert.Equal(destEntryName, edge.Destination.Entry.Name)

		// protocol
		assert.Equal(Protocol, edge.Source.Protocol)
		assert.Equal(Protocol, edge.Destination.Protocol)

		// count
		assert.Equal(count, edge.Count)
	}

	for i, v := range edges {
		if v.Source.Entry.Name == a && v.Destination.Entry.Name == b {
			validateEdge(v, a, b, 2)
			abEdge = i
			continue
		}
		if v.Source.Entry.Name == UnresolvedNodeName && v.Destination.Entry.Name == a {
			validateEdge(v, UnresolvedNodeName, a, 1)
			uaEdge = i
			continue
		}
		if v.Source.Entry.Name == b && v.Destination.Entry.Name == UnresolvedNodeName {
			validateEdge(v, b, UnresolvedNodeName, 1)
			buEdge = i
			continue
		}
		if v.Source.Entry.Name == c && v.Destination.Entry.Name == d {
			validateEdge(v, c, d, 1)
			cdEdge = i
			continue
		}
		if v.Source.Entry.Name == a && v.Destination.Entry.Name == c {
			validateEdge(v, a, c, 1)
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
