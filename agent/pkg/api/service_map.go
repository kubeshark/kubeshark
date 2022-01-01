package api

import (
	"fmt"
	"sync"

	"github.com/up9inc/mizu/shared/logger"
)

const UnresolvedNode = "unresolved"

var instance *graph
var once sync.Once

func GetServiceMapInstance() ServiceMapGraph {
	once.Do(func() {
		instance = newDirectedGraph()
		logger.Log.Debug("Service Map Initialized: %s")
	})
	return instance
}

type ServiceMapGraph interface {
	AddEdge(u, v id, protocol string)
	GetNodes() []string
	PrintNodes()
	PrintAdjacentEdges()
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

func (g *graph) addNode(id id) {
	if _, ok := g.Nodes[id]; ok {
		return
	}
	g.Nodes[id] = struct{}{}
}

func (g *graph) AddEdge(u, v id, p string) {
	if len(u) == 0 {
		u = UnresolvedNode
	}
	if len(v) == 0 {
		v = UnresolvedNode
	}

	if _, ok := g.Nodes[u]; !ok {
		g.addNode(u)
	}
	if _, ok := g.Nodes[v]; !ok {
		g.addNode(v)
	}

	if _, ok := g.Edges[u]; !ok {
		g.Edges[u] = make(map[id]*edgeData)
	}

	if e, ok := g.Edges[u][v]; ok {
		e.count++
	} else {
		g.Edges[u][v] = &edgeData{
			protocol: p,
			count:    1,
		}
	}
}

func (g *graph) GetNodes() []string {
	nodes := make([]string, 0)
	for k := range g.Nodes {
		nodes = append(nodes, string(k))
	}
	return nodes
}

func (g *graph) PrintNodes() {
	fmt.Println("Printing all nodes...")

	for k := range g.Nodes {
		fmt.Printf("Node: %v\n", k)
	}
}

func (g *graph) PrintAdjacentEdges() {
	fmt.Println("Printing all edges...")
	for u, m := range g.Edges {
		for v := range m {
			// Edge exists from u to v.
			fmt.Printf("Edge: %v -> %v - Count: %v\n", u, v, g.Edges[u][v].count)
		}
	}
}
