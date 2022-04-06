package servicemap

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
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
	ProtocolHttp = &tapApi.Protocol{
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
	ProtocolRedis = &tapApi.Protocol{
		Name:            "redis",
		LongName:        "Redis Serialization Protocol",
		Abbreviation:    "REDIS",
		Macro:           "redis",
		Version:         "3.x",
		BackgroundColor: "#a41e11",
		ForegroundColor: "#ffffff",
		FontSize:        11,
		ReferenceLink:   "https://redis.io/topics/protocol",
		Ports:           []string{"6379"},
		Priority:        3,
	}
)

type ServiceMapDisabledSuite struct {
	suite.Suite

	instance *defaultServiceMap
}

type ServiceMapEnabledSuite struct {
	suite.Suite

	instance *defaultServiceMap
}

func (s *ServiceMapDisabledSuite) SetupTest() {
	s.instance = GetDefaultServiceMapInstance()
}

func (s *ServiceMapEnabledSuite) SetupTest() {
	s.instance = GetDefaultServiceMapInstance()
	s.instance.Enable()
}

func (s *ServiceMapDisabledSuite) TestServiceMapInstance() {
	assert := s.Assert()

	assert.NotNil(s.instance)
}

func (s *ServiceMapDisabledSuite) TestServiceMapSingletonInstance() {
	assert := s.Assert()

	instance2 := GetDefaultServiceMapInstance()

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

	s.instance.NewTCPEntry(TCPEntryA, TCPEntryB, ProtocolHttp)
	s.instance.NewTCPEntry(TCPEntryC, TCPEntryD, ProtocolHttp)
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

	// A -> B - HTTP
	s.instance.NewTCPEntry(TCPEntryA, TCPEntryB, ProtocolHttp)

	nodes := s.instance.GetNodes()
	edges := s.instance.GetEdges()

	// Counts for the first entry
	assert.Equal(1, s.instance.GetEntriesProcessedCount())
	assert.Equal(2, s.instance.GetNodesCount())
	assert.Equal(2, len(nodes))
	assert.Equal(1, s.instance.GetEdgesCount())
	assert.Equal(1, len(edges))
	//http protocol
	assert.Equal(1, edges[0].Count)
	assert.Equal(ProtocolHttp.Name, edges[0].Protocol.Name)

	// same A -> B - HTTP, http protocol count should be 2, edges count should be 1
	s.instance.NewTCPEntry(TCPEntryA, TCPEntryB, ProtocolHttp)

	nodes = s.instance.GetNodes()
	edges = s.instance.GetEdges()

	// Counts for a second entry
	assert.Equal(2, s.instance.GetEntriesProcessedCount())
	assert.Equal(2, s.instance.GetNodesCount())
	assert.Equal(2, len(nodes))
	// edges count should still be 1, but http protocol count should be 2
	assert.Equal(1, s.instance.GetEdgesCount())
	assert.Equal(1, len(edges))
	// http protocol
	assert.Equal(2, edges[0].Count) //http
	assert.Equal(ProtocolHttp.Name, edges[0].Protocol.Name)

	// same A -> B - REDIS, http protocol count should be 2 and redis protocol count should 1, edges count should be 2
	s.instance.NewTCPEntry(TCPEntryA, TCPEntryB, ProtocolRedis)

	nodes = s.instance.GetNodes()
	edges = s.instance.GetEdges()

	// Counts after second entry
	assert.Equal(3, s.instance.GetEntriesProcessedCount())
	assert.Equal(2, s.instance.GetNodesCount())
	assert.Equal(2, len(nodes))
	// edges count should be 2, http protocol count should be 2 and redis protocol should be 1
	assert.Equal(2, s.instance.GetEdgesCount())
	assert.Equal(2, len(edges))
	// http and redis protocols
	httpIndex := -1
	redisIndex := -1
	for i, e := range edges {
		if e.Protocol.Name == ProtocolHttp.Name {
			httpIndex = i
			continue
		}
		if e.Protocol.Name == ProtocolRedis.Name {
			redisIndex = i
		}
	}
	assert.NotEqual(-1, httpIndex)
	assert.NotEqual(-1, redisIndex)
	// http protocol
	assert.Equal(2, edges[httpIndex].Count)
	assert.Equal(ProtocolHttp.Name, edges[httpIndex].Protocol.Name)
	// redis protocol
	assert.Equal(1, edges[redisIndex].Count)
	assert.Equal(ProtocolRedis.Name, edges[redisIndex].Protocol.Name)

	// other entries
	s.instance.NewTCPEntry(TCPEntryUnresolved, TCPEntryA, ProtocolHttp)
	s.instance.NewTCPEntry(TCPEntryB, TCPEntryUnresolved2, ProtocolHttp)
	s.instance.NewTCPEntry(TCPEntryC, TCPEntryD, ProtocolHttp)
	s.instance.NewTCPEntry(TCPEntryA, TCPEntryC, ProtocolHttp)

	status := s.instance.GetStatus()
	nodes = s.instance.GetNodes()
	edges = s.instance.GetEdges()
	expectedEntriesProcessedCount := 7
	expectedNodeCount := 6
	expectedEdgeCount := 6

	// Counts after all entries
	assert.Equal(expectedEntriesProcessedCount, s.instance.GetEntriesProcessedCount())
	assert.Equal(expectedNodeCount, s.instance.GetNodesCount())
	assert.Equal(expectedNodeCount, len(nodes))
	assert.Equal(expectedEdgeCount, s.instance.GetEdgesCount())
	assert.Equal(expectedEdgeCount, len(edges))

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
	var validateNode = func(node ServiceMapNode, entryName string, count int) int {
		// id
		assert.GreaterOrEqual(node.Id, 1)
		assert.LessOrEqual(node.Id, expectedNodeCount)

		// entry
		// node.Name is the key of the node, key = entry.Name by default
		// entry.Name is the name of the service and could be unresolved
		// when entry.Name is unresolved, key = entry.IP
		if node.Entry.Name == UnresolvedNodeName {
			assert.Equal(node.Name, node.Entry.IP)
		} else {
			assert.Equal(node.Name, node.Entry.Name)
		}
		assert.Equal(Port, node.Entry.Port)
		assert.Equal(entryName, node.Entry.Name)

		// count
		assert.Equal(count, node.Count)

		return node.Id
	}

	for _, v := range nodes {
		if strings.HasSuffix(v.Name, a) {
			aNode = validateNode(v, a, 5)
			continue
		}
		if strings.HasSuffix(v.Name, b) {
			bNode = validateNode(v, b, 4)
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
	var validateEdge = func(edge ServiceMapEdge, sourceEntryName string, destEntryName string, protocolName string, protocolCount int) {
		// source node
		assert.Contains(nodeIds, edge.Source.Id)
		assert.LessOrEqual(edge.Source.Id, expectedNodeCount)
		if edge.Source.Entry.Name == UnresolvedNodeName {
			assert.Equal(edge.Source.Name, edge.Source.Entry.IP)
		} else {
			assert.Equal(edge.Source.Name, edge.Source.Entry.Name)
		}
		assert.Equal(sourceEntryName, edge.Source.Entry.Name)

		// destination node
		assert.Contains(nodeIds, edge.Destination.Id)
		assert.LessOrEqual(edge.Destination.Id, expectedNodeCount)
		if edge.Destination.Entry.Name == UnresolvedNodeName {
			assert.Equal(edge.Destination.Name, edge.Destination.Entry.IP)
		} else {
			assert.Equal(edge.Destination.Name, edge.Destination.Entry.Name)
		}
		assert.Equal(destEntryName, edge.Destination.Entry.Name)

		// protocol
		assert.Equal(protocolName, edge.Protocol.Name)
		assert.Equal(protocolCount, edge.Count)
	}

	for i, v := range edges {
		if v.Source.Entry.Name == a && v.Destination.Entry.Name == b && v.Protocol.Name == "http" {
			validateEdge(v, a, b, ProtocolHttp.Name, 2)
			abEdge = i
			continue
		}
		if v.Source.Entry.Name == a && v.Destination.Entry.Name == b && v.Protocol.Name == "redis" {
			validateEdge(v, a, b, ProtocolRedis.Name, 1)
			abEdge = i
			continue
		}
		if v.Source.Entry.Name == UnresolvedNodeName && v.Destination.Entry.Name == a {
			validateEdge(v, UnresolvedNodeName, a, ProtocolHttp.Name, 1)
			uaEdge = i
			continue
		}
		if v.Source.Entry.Name == b && v.Destination.Entry.Name == UnresolvedNodeName {
			validateEdge(v, b, UnresolvedNodeName, ProtocolHttp.Name, 1)
			buEdge = i
			continue
		}
		if v.Source.Entry.Name == c && v.Destination.Entry.Name == d {
			validateEdge(v, c, d, ProtocolHttp.Name, 1)
			cdEdge = i
			continue
		}
		if v.Source.Entry.Name == a && v.Destination.Entry.Name == c {
			validateEdge(v, a, c, ProtocolHttp.Name, 1)
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
	assert.Equal([]ServiceMapNode{}, nodes)

	// Edges after reset
	assert.Equal([]ServiceMapEdge{}, edges)
}

func TestServiceMapSuite(t *testing.T) {
	suite.Run(t, new(ServiceMapDisabledSuite))
	suite.Run(t, new(ServiceMapEnabledSuite))
}
