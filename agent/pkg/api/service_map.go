package api

import (
	"sync"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

const (
	ServiceMapEnabled  = "enabled"
	ServiceMapDisabled = "disabled"
	UnresolvedNode     = "unresolved"
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
	AddEdge(source, destination key, protocol string)
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

type nodeData struct {
	id       int
	protocol string
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

func newNodeData(id int, p string) *nodeData {
	return &nodeData{
		id:       id,
		protocol: p,
		count:    1,
	}
}

func newEdgeData() *edgeData {
	return &edgeData{
		count: 1,
	}
}

func (s *serviceMap) addNode(k key, p string) {
	if _, ok := s.graph.Nodes[k]; ok {
		return
	}
	s.graph.Nodes[k] = newNodeData(len(s.graph.Nodes)+1, p)
}

func (s *serviceMap) AddEdge(u, v key, p string) {
	if !s.IsEnabled() {
		return
	}

	if len(u) == 0 {
		u = UnresolvedNode
	}
	if len(v) == 0 {
		v = UnresolvedNode
	}

	if n, ok := s.graph.Nodes[u]; !ok {
		s.addNode(u, p)
	} else {
		n.count++
	}
	if n, ok := s.graph.Nodes[v]; !ok {
		s.addNode(v, p)
	} else {
		n.count++
	}

	if _, ok := s.graph.Edges[u]; !ok {
		s.graph.Edges[u] = make(map[key]*edgeData)
	}

	if e, ok := s.graph.Edges[u][v]; ok {
		e.count++
	} else {
		s.graph.Edges[u][v] = newEdgeData()
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
			Name:     string(i),
			Id:       n.id,
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
					Name:     string(u),
					Id:       s.graph.Nodes[u].id,
					Protocol: s.graph.Nodes[u].protocol,
					Count:    s.graph.Nodes[u].count,
				},
				Destination: shared.ServiceMapNode{
					Name:     string(v),
					Id:       s.graph.Nodes[v].id,
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
