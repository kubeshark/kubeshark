package controllers

import (
	"mizuserver/pkg/api"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/up9inc/mizu/shared"
)

func calcNodeSize(value float32) float32 {
	//TODO: Linear interpolation
	if value < 1 {
		value = 1
	}
	if value > 300 {
		value = 300
	}
	return value
}

func buildGraph() *components.Page {
	page := components.NewPage()
	page.SetLayout(components.PageFlexLayout)
	page.PageTitle = "Mizu Service Map"

	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: "MIZU",
			Width:     "1800px",
			Height:    "1000px",
			//Theme:     types.ThemeInfographic,
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "Service Map Graph",
		}))

	// TODO: Sort nodes by name
	// TODO: Add protocol color

	var graphNodes []opts.GraphNode
	for _, n := range api.GetServiceMapInstance().GetNodes() {
		graphNodes = append(graphNodes, opts.GraphNode{
			Name:       n.Name,
			Value:      float32(n.Count),
			Fixed:      true,
			SymbolSize: calcNodeSize(float32(n.Count)),
		})
	}

	var graphEdges []opts.GraphLink
	for _, e := range api.GetServiceMapInstance().GetEdges() {
		graphEdges = append(graphEdges, opts.GraphLink{
			Source: e.Source,
			Target: e.Destination,
			Value:  float32(e.Count),
		})
	}

	graph.AddSeries("graph", graphNodes, graphEdges).
		SetSeriesOptions(
			charts.WithGraphChartOpts(opts.GraphChart{
				Layout:             "circular",
				Force:              &opts.GraphForce{Repulsion: 8000},
				Roam:               true,
				FocusNodeAdjacency: true,
			}),
			charts.WithEmphasisOpts(opts.Emphasis{
				Label: &opts.Label{
					Show:     true,
					Color:    "red",
					Position: "top",
				},
			}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Curveness: 0.3,
				Width:     2,
				Type:      "dotted",
			}),
			charts.WithLabelOpts(opts.Label{Show: true, Position: "top"}),
		)

	return page.AddCharts(graph)
}

type ServiceMapController struct{}

func NewServiceMapController() *ServiceMapController {
	return &ServiceMapController{}
}

func (s *ServiceMapController) Status(c *gin.Context) {
	serviceMap := api.GetServiceMapInstance()
	status := &shared.ServiceMapStatus{
		Status:                "enabled",
		EntriesProcessedCount: serviceMap.GetEntriesProcessedCount(),
		NodeCount:             serviceMap.GetNodesCount(),
		EdgeCount:             serviceMap.GetEdgesCount(),
	}
	c.JSON(http.StatusOK, status)
}

func (s *ServiceMapController) Get(c *gin.Context) {
	serviceMap := api.GetServiceMapInstance()
	response := &shared.ServiceMapResponse{
		Nodes: serviceMap.GetNodes(),
		Edges: serviceMap.GetEdges(),
	}
	c.JSON(http.StatusOK, response)
}

func (s *ServiceMapController) Render(c *gin.Context) {
	w := c.Writer
	header := w.Header()
	header.Set("Content-Type", "text/html")

	err := buildGraph().Render(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.(http.Flusher).Flush()
}

func (s *ServiceMapController) Reset(c *gin.Context) {
	serviceMap := api.GetServiceMapInstance()
	serviceMap.Reset()
	s.Status(c)
}
