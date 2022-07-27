package servicemap

import (
	"sync"

	"github.com/jinzhu/copier"

	"github.com/up9inc/mizu/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

const (
	ServiceMapEnabled  = "enabled"
	ServiceMapDisabled = "disabled"
	UnresolvedNodeName = "unresolved"
)

var instance *defaultServiceMap
var once sync.Once

func GetDefaultServiceMapInstance() *defaultServiceMap {
	once.Do(func() {
		instance = NewDefaultServiceMapGenerator()
		logger.Log.Debug("Service Map Initialized")
	})
	return instance
}

type defaultServiceMap struct {
	enabled          bool
	graph            *graph
	entriesProcessed int
}

type ServiceMapSink interface {
	NewTCPEntry(source *tapApi.TCP, destination *tapApi.TCP, protocol *tapApi.Protocol)
}

type ServiceMap interface {
	Enable()
	Disable()
	IsEnabled() bool
	GetStatus() ServiceMapStatus
	GetNodes() []ServiceMapNode
	GetEdges() []ServiceMapEdge
	GetEntriesProcessedCount() int
	GetNodesCount() int
	GetEdgesCount() int
	Reset()
}

func NewDefaultServiceMapGenerator() *defaultServiceMap {
	return &defaultServiceMap{
		enabled:          false,
		entriesProcessed: 0,
		graph:            newDirectedGraph(),
	}
}

type key string

type entryData struct {
	key   key
	entry *tapApi.TCP
}

type nodeData struct {
	id    int
	entry *tapApi.TCP
	count int
}

type edgeProtocol struct {
	protocol *tapApi.Protocol
	count    int
}

type edgeData struct {
	data map[key]*edgeProtocol
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

func newNodeData(id int, e *tapApi.TCP) *nodeData {
	return &nodeData{
		id:    id,
		entry: e,
		count: 1,
	}
}

func newEdgeData(p *tapApi.Protocol) *edgeData {
	return &edgeData{
		data: map[key]*edgeProtocol{
			key(p.Name): {
				protocol: p,
				count:    1,
			},
		},
	}
}

func (s *defaultServiceMap) nodeExists(k key) (*nodeData, bool) {
	n, ok := s.graph.Nodes[k]
	return n, ok
}

func (s *defaultServiceMap) addNode(k key, e *tapApi.TCP) (*nodeData, bool) {
	nd, exists := s.nodeExists(k)
	if !exists {
		s.graph.Nodes[k] = newNodeData(len(s.graph.Nodes)+1, e)
		return s.graph.Nodes[k], true
	}
	return nd, false
}

func (s *defaultServiceMap) addEdge(u, v *entryData, p *tapApi.Protocol) {
	if n, ok := s.addNode(u.key, u.entry); !ok {
		n.count++
	}
	if n, ok := s.addNode(v.key, v.entry); !ok {
		n.count++
	}

	if _, ok := s.graph.Edges[u.key]; !ok {
		s.graph.Edges[u.key] = make(map[key]*edgeData)
	}

	// new edge u -> v pair
	// protocol is the same for u and v
	if e, ok := s.graph.Edges[u.key][v.key]; ok {
		// edge data already exists for u -> v pair
		// we have a new protocol for this u -> v pair

		k := key(p.Name)
		if pd, pOk := e.data[k]; pOk {
			// protocol key already exists, just increment the count
			pd.count++
		} else {
			// new protocol key
			e.data[k] = &edgeProtocol{
				protocol: p,
				count:    1,
			}
		}
	} else {
		// new edge data for u -> v pair
		s.graph.Edges[u.key][v.key] = newEdgeData(p)
	}

	s.entriesProcessed++
}

func (s *defaultServiceMap) Enable() {
	s.enabled = true
}

func (s *defaultServiceMap) Disable() {
	s.Reset()
	s.enabled = false
}

func (s *defaultServiceMap) IsEnabled() bool {
	return s.enabled
}

func (s *defaultServiceMap) NewTCPEntry(src *tapApi.TCP, dst *tapApi.TCP, p *tapApi.Protocol) {
	if !s.IsEnabled() {
		return
	}

	var srcEntry *entryData
	var dstEntry *entryData

	if len(src.Name) == 0 {
		srcEntry = &entryData{
			key:   key(src.IP),
			entry: &tapApi.TCP{},
		}
		if err := copier.Copy(srcEntry.entry, src); err != nil {
			logger.Log.Errorf("Error while copying src entry into src entry data")
		}

		srcEntry.entry.Name = UnresolvedNodeName
	} else {
		srcEntry = &entryData{
			key:   key(src.Name),
			entry: src,
		}
	}

	if len(dst.Name) == 0 {
		dstEntry = &entryData{
			key:   key(dst.IP),
			entry: &tapApi.TCP{},
		}
		if err := copier.Copy(dstEntry.entry, dst); err != nil {
			logger.Log.Errorf("Error while copying dst entry into dst entry data")
		}

		dstEntry.entry.Name = UnresolvedNodeName
	} else {
		dstEntry = &entryData{
			key:   key(dst.Name),
			entry: dst,
		}
	}

	s.addEdge(srcEntry, dstEntry, p)
}

func (s *defaultServiceMap) GetStatus() ServiceMapStatus {
	status := ServiceMapDisabled
	if s.IsEnabled() {
		status = ServiceMapEnabled
	}

	return ServiceMapStatus{
		Status:                status,
		EntriesProcessedCount: s.entriesProcessed,
		NodeCount:             s.GetNodesCount(),
		EdgeCount:             s.GetEdgesCount(),
	}
}

func (s *defaultServiceMap) GetNodes() []ServiceMapNode {
	nodes := []ServiceMapNode{}

	for i, n := range s.graph.Nodes {
		nodes = append(nodes, ServiceMapNode{
			Id:       n.id,
			Name:     string(i),
			Resolved: n.entry.Name != UnresolvedNodeName,
			Entry:    n.entry,
			Count:    n.count,
		})
	}

	return nodes
}

func (s *defaultServiceMap) GetEdges() []ServiceMapEdge {
	edges := []ServiceMapEdge{}

	for u, m := range s.graph.Edges {
		for v := range m {
			for _, p := range s.graph.Edges[u][v].data {
				edges = append(edges, ServiceMapEdge{
					Source: ServiceMapNode{
						Id:       s.graph.Nodes[u].id,
						Name:     string(u),
						Entry:    s.graph.Nodes[u].entry,
						Resolved: s.graph.Nodes[u].entry.Name != UnresolvedNodeName,
						Count:    s.graph.Nodes[u].count,
					},
					Destination: ServiceMapNode{
						Id:       s.graph.Nodes[v].id,
						Name:     string(v),
						Entry:    s.graph.Nodes[v].entry,
						Resolved: s.graph.Nodes[v].entry.Name != UnresolvedNodeName,
						Count:    s.graph.Nodes[v].count,
					},
					Count:    p.count,
					Protocol: p.protocol,
				})
			}
		}
	}

	return edges
}

func (s *defaultServiceMap) GetEntriesProcessedCount() int {
	return s.entriesProcessed
}

func (s *defaultServiceMap) GetNodesCount() int {
	return len(s.graph.Nodes)
}

func (s *defaultServiceMap) GetEdgesCount() int {
	var count int
	for u, m := range s.graph.Edges {
		for v := range m {
			for range s.graph.Edges[u][v].data {
				count++
			}
		}
	}
	return count
}

func (s *defaultServiceMap) Reset() {
	s.entriesProcessed = 0
	s.graph = newDirectedGraph()
}
