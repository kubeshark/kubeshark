package api

import (
	"fmt"
	"sync"

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
	AddEdge(source, destination id, protocol string)
	GetNodes() []string
	PrintNodes()
	PrintAdjacentEdges()
	GetEntriesProcessedCount() int
	GetNodesCount() int
	GetEdgesCount() int
}

func newServiceMap() *serviceMap {
	return &serviceMap{
		entriesProcessed: 0,
		graph:            newDirectedGraph(),
	}
}

type id string
type edgeData struct {
	protocol string
	count    int
}

type graph struct {
	Nodes map[id]struct{}
	Edges map[id]map[id]*edgeData
}

func newDirectedGraph() *graph {
	return &graph{
		Nodes: make(map[id]struct{}),
		Edges: make(map[id]map[id]*edgeData),
	}
}

func newEdgeData(p string) *edgeData {
	return &edgeData{
		protocol: p,
		count:    1,
	}
}

func (s *serviceMap) addNode(id id) {
	if _, ok := s.graph.Nodes[id]; ok {
		return
	}
	s.graph.Nodes[id] = struct{}{}
}

func (s *serviceMap) AddEdge(u, v id, p string) {
	if len(u) == 0 {
		u = UnresolvedNode
	}
	if len(v) == 0 {
		v = UnresolvedNode
	}

	if _, ok := s.graph.Nodes[u]; !ok {
		s.addNode(u)
	}
	if _, ok := s.graph.Nodes[v]; !ok {
		s.addNode(v)
	}

	if _, ok := s.graph.Edges[u]; !ok {
		s.graph.Edges[u] = make(map[id]*edgeData)
	}

	if e, ok := s.graph.Edges[u][v]; ok {
		e.count++
	} else {
		s.graph.Edges[u][v] = &edgeData{
			protocol: p,
			count:    1,
		}
	}

	s.entriesProcessed++
}

func (s *serviceMap) GetNodes() []string {
	nodes := make([]string, 0)
	for k := range s.graph.Nodes {
		nodes = append(nodes, string(k))
	}
	return nodes
}

func (s *serviceMap) PrintNodes() {
	fmt.Println("Printing all nodes...")

	for k := range s.graph.Nodes {
		fmt.Printf("Node: %v\n", k)
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
	return len(s.graph.Edges)
}
