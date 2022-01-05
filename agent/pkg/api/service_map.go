package api

import (
	"fmt"
	"sync"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

const UnresolvedNode = "unresolved"

var instance *serviceMap
var once sync.Once

func GetServiceMapInstance() ServiceMap {
	once.Do(func() {
		instance = newServiceMap()
		logger.Log.Debug("Service Map Initialized: %s")
	})
	return instance
}

type serviceMap struct {
	graph            *graph
	entriesProcessed int
}

type ServiceMap interface {
	AddEdge(source, destination key, protocol string)
	GetNodes() []shared.ServiceMapNode
	GetEdges() []shared.ServiceMapEdge
	PrintNodes()
	PrintAdjacentEdges()
	GetEntriesProcessedCount() int
	GetNodesCount() int
	GetEdgesCount() int
	Reset()
}

func newServiceMap() *serviceMap {
	return &serviceMap{
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

func (s *serviceMap) PrintNodes() {
	fmt.Println("Printing all nodes...")

	for k, n := range s.graph.Nodes {
		fmt.Printf("Node: %v Id: %v - Protocol: %v Count: %v\n", k, n.id, n.protocol, n.count)
	}
}

func (s *serviceMap) PrintAdjacentEdges() {
	fmt.Println("Printing all edges...")
	for u, m := range s.graph.Edges {
		for v := range m {
			// Edge exists from u to v.
			fmt.Printf("Edge: %v -> %v - Count: %v\n", u, v, s.graph.Edges[u][v].count)
		}
	}
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
