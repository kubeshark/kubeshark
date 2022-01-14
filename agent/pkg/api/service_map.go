package api

import (
	"sync"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

const (
	ServiceMapEnabled  = "enabled"
	ServiceMapDisabled = "disabled"
	UnresolvedNodeName = "unresolved"
)

var instance *serviceMap
var once sync.Once

func GetServiceMapInstance() ServiceMap {
	once.Do(func() {
		instance = newServiceMap()
		logger.Log.Debug("Service Map Initialized")
	})
	return instance
}

type serviceMap struct {
	config           *shared.MizuAgentConfig
	graph            *graph
	entriesProcessed int
}

type ServiceMap interface {
	SetConfig(config *shared.MizuAgentConfig)
	IsEnabled() bool
	NewTCPEntry(source *tapApi.TCP, destination *tapApi.TCP, protocol *tapApi.Protocol)
	GetStatus() shared.ServiceMapStatus
	GetNodes() []shared.ServiceMapNode
	GetEdges() []shared.ServiceMapEdge
	GetEntriesProcessedCount() int
	GetNodesCount() int
	GetEdgesCount() int
	Reset()
}

func newServiceMap() *serviceMap {
	return &serviceMap{
		config:           nil,
		entriesProcessed: 0,
		graph:            newDirectedGraph(),
	}
}

type key string

type entryData struct {
	key      key
	entry    *tapApi.TCP
	protocol *tapApi.Protocol
}

type nodeData struct {
	id       int
	entry    *tapApi.TCP
	protocol *tapApi.Protocol
	count    int
}
type edgeData struct {
	count int
}

type graph struct {
	Nodes map[key]*nodeData
	Edges map[key]map[key]*edgeData
}

func newDirectedGraph() *graph {
	return &graph{
		Nodes: make(map[key]*nodeData),
		Edges: make(map[key]map[key]*edgeData),
	}
}

func newNodeData(id int, e *tapApi.TCP, p *tapApi.Protocol) *nodeData {
	return &nodeData{
		id:       id,
		entry:    e,
		protocol: p,
		count:    1,
	}
}

func newEdgeData() *edgeData {
	return &edgeData{
		count: 1,
	}
}

func (s *serviceMap) nodeExists(k key) (*nodeData, bool) {
	n, ok := s.graph.Nodes[k]
	return n, ok
}

func (s *serviceMap) addNode(k key, e *tapApi.TCP, p *tapApi.Protocol) (*nodeData, bool) {
	nd, exists := s.nodeExists(k)
	if !exists {
		s.graph.Nodes[k] = newNodeData(len(s.graph.Nodes)+1, e, p)
		return s.graph.Nodes[k], true
	}
	return nd, false
}

func (s *serviceMap) addEdge(u, v *entryData) {
	if n, ok := s.addNode(u.key, u.entry, u.protocol); !ok {
		n.count++
	}
	if n, ok := s.addNode(v.key, v.entry, v.protocol); !ok {
		n.count++
	}

	if _, ok := s.graph.Edges[u.key]; !ok {
		s.graph.Edges[u.key] = make(map[key]*edgeData)
	}

	if e, ok := s.graph.Edges[u.key][v.key]; ok {
		e.count++
	} else {
		s.graph.Edges[u.key][v.key] = newEdgeData()
	}

	s.entriesProcessed++
}

func (s *serviceMap) SetConfig(config *shared.MizuAgentConfig) {
	s.config = config
}

func (s *serviceMap) IsEnabled() bool {
	if s.config != nil && s.config.ServiceMap {
		return true
	}
	return false
}

func (s *serviceMap) NewTCPEntry(src *tapApi.TCP, dst *tapApi.TCP, p *tapApi.Protocol) {
	if !s.IsEnabled() {
		return
	}

	srcEntry := &entryData{
		key:      key(src.IP),
		entry:    src,
		protocol: p,
	}
	if len(srcEntry.entry.Name) == 0 {
		srcEntry.entry.Name = UnresolvedNodeName
	}

	dstEntry := &entryData{
		key:      key(dst.IP),
		entry:    dst,
		protocol: p,
	}
	if len(dstEntry.entry.Name) == 0 {
		dstEntry.entry.Name = UnresolvedNodeName
	}

	s.addEdge(srcEntry, dstEntry)
}

func (s *serviceMap) GetStatus() shared.ServiceMapStatus {
	status := ServiceMapDisabled
	if s.IsEnabled() {
		status = ServiceMapEnabled
	}

	return shared.ServiceMapStatus{
		Status:                status,
		EntriesProcessedCount: s.entriesProcessed,
		NodeCount:             s.GetNodesCount(),
		EdgeCount:             s.GetEdgesCount(),
	}
}

func (s *serviceMap) GetNodes() []shared.ServiceMapNode {
	var nodes []shared.ServiceMapNode
	for i, n := range s.graph.Nodes {
		nodes = append(nodes, shared.ServiceMapNode{
			Id:       n.id,
			Name:     string(i),
			Entry:    n.entry,
			Protocol: n.protocol,
			Count:    n.count,
		})
	}
	return nodes
}

func (s *serviceMap) GetEdges() []shared.ServiceMapEdge {
	var edges []shared.ServiceMapEdge
	for u, m := range s.graph.Edges {
		for v := range m {
			edges = append(edges, shared.ServiceMapEdge{
				Source: shared.ServiceMapNode{
					Id:       s.graph.Nodes[u].id,
					Name:     string(u),
					Entry:    s.graph.Nodes[u].entry,
					Protocol: s.graph.Nodes[u].protocol,
					Count:    s.graph.Nodes[u].count,
				},
				Destination: shared.ServiceMapNode{
					Id:       s.graph.Nodes[v].id,
					Name:     string(v),
					Entry:    s.graph.Nodes[v].entry,
					Protocol: s.graph.Nodes[v].protocol,
					Count:    s.graph.Nodes[v].count,
				},
				Count: s.graph.Edges[u][v].count,
			})
		}
	}
	return edges
}

func (s *serviceMap) GetEntriesProcessedCount() int {
	return s.entriesProcessed
}

func (s *serviceMap) GetNodesCount() int {
	return len(s.graph.Nodes)
}

func (s *serviceMap) GetEdgesCount() int {
	var count int
	for _, m := range s.graph.Edges {
		for range m {
			count++
		}
	}
	return count
}

func (s *serviceMap) Reset() {
	s.entriesProcessed = 0
	s.graph = newDirectedGraph()
}
