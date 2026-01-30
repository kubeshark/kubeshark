package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/rs/zerolog"
)

// MCP Protocol types (JSON-RPC 2.0)

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Result  any            `json:"result,omitempty"`
	Error   *jsonRPCError  `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MCP-specific types

type mcpInitializeParams struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    json.RawMessage `json:"capabilities"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

type mcpInitializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    struct {
		Tools struct{} `json:"tools"`
	} `json:"capabilities"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type mcpTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type mcpListToolsResult struct {
	Tools []mcpTool `json:"tools"`
}

type mcpCallToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type mcpCallToolResult struct {
	Content []mcpContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MCP Server

type mcpServer struct {
	hubBaseURL string
	httpClient *http.Client
	stdin      io.Reader
	stdout     io.Writer
}

func runMCP() {
	// Disable zerolog output to stderr (MCP uses stdio)
	zerolog.SetGlobalLevel(zerolog.Disabled)

	kubernetesProvider, err := getKubernetesProviderForCli(true, true)
	if err != nil {
		writeErrorToStderr("Failed to get Kubernetes provider: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Check if Kubeshark services exist
	exists, err := kubernetesProvider.DoesServiceExist(ctx, config.Config.Tap.Release.Namespace, kubernetes.FrontServiceName)
	if err != nil || !exists {
		writeErrorToStderr("Kubeshark front service not found. Run '%s tap' first.", misc.Program)
		return
	}

	// Start proxy to frontend
	frontURL := kubernetes.GetProxyOnPort(config.Config.Tap.Proxy.Front.Port)
	response, err := http.Get(fmt.Sprintf("%s/", frontURL))
	if err != nil || response.StatusCode != 200 {
		startProxyReportErrorIfAny(
			kubernetesProvider,
			ctx,
			kubernetes.FrontServiceName,
			kubernetes.FrontPodName,
			configStructs.ProxyFrontPortLabel,
			config.Config.Tap.Proxy.Front.Port,
			configStructs.ContainerPort,
			"",
		)
		connector := connect.NewConnector(frontURL, connect.DefaultRetries, connect.DefaultTimeout)
		if err := connector.TestConnection(""); err != nil {
			writeErrorToStderr("Couldn't connect to Kubeshark frontend")
			return
		}
	}

	// Hub MCP API is available via frontend at /mcp/*
	hubMCPURL := fmt.Sprintf("%s/mcp", frontURL)

	server := &mcpServer{
		hubBaseURL: hubMCPURL,
		httpClient: &http.Client{},
		stdin:      os.Stdin,
		stdout:     os.Stdout,
	}

	server.run()
}

func writeErrorToStderr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (s *mcpServer) run() {
	scanner := bufio.NewScanner(s.stdin)
	// Increase buffer size for large messages
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		s.handleRequest(&req)
	}
}

func (s *mcpServer) handleRequest(req *jsonRPCRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "initialized":
		// Notification, no response needed
	case "tools/list":
		s.handleListTools(req)
	case "tools/call":
		s.handleCallTool(req)
	case "ping":
		s.sendResult(req.ID, map[string]any{})
	default:
		s.sendError(req.ID, -32601, "Method not found", req.Method)
	}
}

func (s *mcpServer) handleInitialize(req *jsonRPCRequest) {
	result := mcpInitializeResult{
		ProtocolVersion: "2024-11-05",
	}
	result.ServerInfo.Name = "kubeshark-mcp"
	result.ServerInfo.Version = "1.0.0"

	s.sendResult(req.ID, result)
}

func (s *mcpServer) handleListTools(req *jsonRPCRequest) {
	tools := []mcpTool{
		{
			Name:        "list_workloads",
			Description: "List Kubernetes workloads (pods, services, namespaces, nodes) with L7 API traffic observed by Kubeshark",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"type": {
						"type": "string",
						"description": "Workload type: pod, service, namespace, or node",
						"enum": ["pod", "service", "namespace", "node"],
						"default": "pod"
					},
					"ns": {
						"type": "string",
						"description": "Filter by namespace"
					},
					"labels": {
						"type": "string",
						"description": "Filter by labels (format: key=value,key2=value2)"
					}
				}
			}`),
		},
		{
			Name:        "list_api_calls",
			Description: "Query L7 API transactions (HTTP, gRPC, etc.) captured by Kubeshark",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"src_ns": {
						"type": "string",
						"description": "Filter by source namespace(s), comma-separated"
					},
					"src_pod": {
						"type": "string",
						"description": "Filter by source pod name (prefix match)"
					},
					"dst_ns": {
						"type": "string",
						"description": "Filter by destination namespace(s), comma-separated"
					},
					"dst_pod": {
						"type": "string",
						"description": "Filter by destination pod name (prefix match)"
					},
					"dst_svc": {
						"type": "string",
						"description": "Filter by destination service name"
					},
					"proto": {
						"type": "string",
						"description": "Filter by protocol (http, grpc, redis, kafka, etc.)"
					},
					"method": {
						"type": "string",
						"description": "Filter by HTTP method or gRPC method"
					},
					"path": {
						"type": "string",
						"description": "Filter by request path (prefix match)"
					},
					"status": {
						"type": "string",
						"description": "Filter by status: ok, error, or specific code"
					},
					"limit": {
						"type": "integer",
						"description": "Maximum number of results (default: 100)",
						"default": 100
					}
				}
			}`),
		},
		{
			Name:        "get_api_call",
			Description: "Get detailed information about a specific API call by ID",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {
						"type": "string",
						"description": "The API call ID"
					},
					"include_headers": {
						"type": "boolean",
						"description": "Include request/response headers",
						"default": false
					},
					"include_payload": {
						"type": "boolean",
						"description": "Include request/response payload (truncated to 4KB)",
						"default": false
					}
				},
				"required": ["id"]
			}`),
		},
		{
			Name:        "get_api_stats",
			Description: "Get aggregated statistics for API calls",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"ns": {
						"type": "string",
						"description": "Filter by namespace"
					},
					"svc": {
						"type": "string",
						"description": "Filter by service"
					},
					"pod": {
						"type": "string",
						"description": "Filter by pod"
					},
					"group_by": {
						"type": "string",
						"description": "Group results by: endpoint, status, src, dst, proto",
						"enum": ["endpoint", "status", "src", "dst", "proto"],
						"default": "endpoint"
					}
				}
			}`),
		},
	}

	s.sendResult(req.ID, mcpListToolsResult{Tools: tools})
}

func (s *mcpServer) handleCallTool(req *jsonRPCRequest) {
	var params mcpCallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	var result string
	var isError bool

	switch params.Name {
	case "list_workloads":
		result, isError = s.callListWorkloads(params.Arguments)
	case "list_api_calls":
		result, isError = s.callListAPICalls(params.Arguments)
	case "get_api_call":
		result, isError = s.callGetAPICall(params.Arguments)
	case "get_api_stats":
		result, isError = s.callGetAPIStats(params.Arguments)
	default:
		s.sendError(req.ID, -32602, "Unknown tool", params.Name)
		return
	}

	s.sendResult(req.ID, mcpCallToolResult{
		Content: []mcpContent{{Type: "text", Text: result}},
		IsError: isError,
	})
}

func (s *mcpServer) callListWorkloads(args map[string]any) (string, bool) {
	query := url.Values{}
	if v, ok := args["type"].(string); ok && v != "" {
		query.Set("type", v)
	}
	if v, ok := args["ns"].(string); ok && v != "" {
		query.Set("ns", v)
	}
	if v, ok := args["labels"].(string); ok && v != "" {
		query.Set("labels", v)
	}

	return s.doGet("/workloads", query)
}

func (s *mcpServer) callListAPICalls(args map[string]any) (string, bool) {
	query := url.Values{}
	for _, key := range []string{"src_ns", "src_pod", "dst_ns", "dst_pod", "dst_svc", "proto", "method", "path", "status"} {
		if v, ok := args[key].(string); ok && v != "" {
			query.Set(key, v)
		}
	}
	if v, ok := args["limit"].(float64); ok {
		query.Set("limit", fmt.Sprintf("%d", int(v)))
	}

	return s.doGet("/calls", query)
}

func (s *mcpServer) callGetAPICall(args map[string]any) (string, bool) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "Error: id is required", true
	}

	query := url.Values{}
	if v, ok := args["include_headers"].(bool); ok && v {
		query.Set("include_headers", "true")
	}
	if v, ok := args["include_payload"].(bool); ok && v {
		query.Set("include_payload", "true")
	}

	return s.doGet("/calls/"+url.PathEscape(id), query)
}

func (s *mcpServer) callGetAPIStats(args map[string]any) (string, bool) {
	query := url.Values{}
	for _, key := range []string{"ns", "svc", "pod", "group_by"} {
		if v, ok := args[key].(string); ok && v != "" {
			query.Set(key, v)
		}
	}

	return s.doGet("/stats", query)
}

func (s *mcpServer) doGet(path string, query url.Values) (string, bool) {
	reqURL := s.hubBaseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	resp, err := s.httpClient.Get(reqURL)
	if err != nil {
		return fmt.Sprintf("Error calling Hub API: %v", err), true
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err), true
	}

	if resp.StatusCode >= 400 {
		return fmt.Sprintf("Hub API error (%d): %s", resp.StatusCode, string(body)), true
	}

	// Pretty-print JSON for readability
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		return string(body), false
	}
	return prettyJSON.String(), false
}

func (s *mcpServer) sendResult(id any, result any) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.send(resp)
}

func (s *mcpServer) sendError(id any, code int, message string, data any) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &jsonRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.send(resp)
}

func (s *mcpServer) send(resp jsonRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		writeErrorToStderr("Failed to marshal response: %v", err)
		return
	}
	fmt.Fprintln(s.stdout, string(data))
}
