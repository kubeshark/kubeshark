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
	"os/exec"
	"strings"
	"sync"
	"time"

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
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MCP-specific types

type mcpInitializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    struct {
		Tools   struct{} `json:"tools"`
		Prompts struct{} `json:"prompts"`
	} `json:"capabilities"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
	Instructions string `json:"instructions,omitempty"`
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

type mcpPrompt struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type mcpListPromptsResult struct {
	Prompts []mcpPrompt `json:"prompts"`
}

type mcpGetPromptParams struct {
	Name string `json:"name"`
}

type mcpPromptMessage struct {
	Role    string     `json:"role"`
	Content mcpContent `json:"content"`
}

type mcpGetPromptResult struct {
	Messages []mcpPromptMessage `json:"messages"`
}

// MCP Server

type mcpServer struct {
	hubBaseURL         string
	httpClient         *http.Client
	stdin              io.Reader
	stdout             io.Writer
	backendInitialized bool
	backendMu          sync.Mutex
	tapSetFlags        []string // Flags to pass to 'kubeshark tap' when starting
}

func runMCPWithConfig(tapSetFlags []string) {
	// Disable zerolog output to stderr (MCP uses stdio)
	zerolog.SetGlobalLevel(zerolog.Disabled)

	// Initialize the MCP server without requiring the backend to be running
	// Backend connection will be established lazily when API tools are called
	server := &mcpServer{
		httpClient:  &http.Client{},
		stdin:       os.Stdin,
		stdout:      os.Stdout,
		tapSetFlags: tapSetFlags,
	}

	server.run()
}

// ensureBackendConnection establishes connection to Kubeshark backend if not already connected
// Returns an error message if connection fails, empty string on success
func (s *mcpServer) ensureBackendConnection() string {
	s.backendMu.Lock()
	defer s.backendMu.Unlock()

	if s.backendInitialized {
		return ""
	}

	kubernetesProvider, err := getKubernetesProviderForCli(true, true)
	if err != nil {
		return fmt.Sprintf("Failed to get Kubernetes provider: %v. Is kubectl configured?", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Check if Kubeshark services exist
	exists, err := kubernetesProvider.DoesServiceExist(ctx, config.Config.Tap.Release.Namespace, kubernetes.FrontServiceName)
	if err != nil || !exists {
		return "Kubeshark is not running. Use the 'start_kubeshark' tool to start it first."
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
			return "Couldn't connect to Kubeshark frontend. Is Kubeshark running?"
		}
	}

	// Hub MCP API is available via frontend at /api/mcp/*
	s.hubBaseURL = fmt.Sprintf("%s/api/mcp", frontURL)
	s.backendInitialized = true
	return ""
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
	case "initialized", "notifications/initialized":
		// Notification, no response needed
		return
	case "tools/list":
		s.handleListTools(req)
	case "tools/call":
		s.handleCallTool(req)
	case "prompts/list":
		s.handleListPrompts(req)
	case "prompts/get":
		s.handleGetPrompt(req)
	case "ping":
		s.sendResult(req.ID, map[string]any{})
	default:
		s.sendError(req.ID, -32601, "Method not found", req.Method)
	}
}

func (s *mcpServer) handleInitialize(req *jsonRPCRequest) {
	result := mcpInitializeResult{
		ProtocolVersion: "2024-11-05",
		Instructions: `When working with Kubeshark, ALWAYS use the provided MCP tools instead of kubectl or helm commands:
- To check if Kubeshark is running: use 'check_kubeshark_status' (NOT kubectl get pods)
- To start Kubeshark: use 'start_kubeshark' (NOT kubectl apply or helm install)
- To stop Kubeshark: use 'stop_kubeshark' (NOT kubectl delete or helm uninstall)
- To query traffic data: use 'list_workloads', 'list_api_calls', 'get_api_call', 'get_api_stats'
These tools provide proper integration and accurate results.`,
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
					},
					"group_by": {
						"type": "string",
						"description": "Group results by: node, ns (namespace), or worker",
						"enum": ["node", "ns", "worker"]
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
		{
			Name:        "start_kubeshark",
			Description: "REQUIRED: Use this tool to start/run/deploy Kubeshark for capturing network traffic. Do NOT use kubectl or helm directly - this tool handles the deployment correctly with all required settings.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespaces": {
						"type": "string",
						"description": "Comma-separated list of namespaces to tap (e.g., 'default,kube-system'). If not specified, taps all namespaces."
					},
					"pod_regex": {
						"type": "string",
						"description": "Regular expression to filter pods by name (e.g., 'nginx.*')"
					},
					"release_namespace": {
						"type": "string",
						"description": "Namespace where Kubeshark will be installed (default: 'default')"
					}
				}
			}`),
		},
		{
			Name:        "stop_kubeshark",
			Description: "REQUIRED: Use this tool to stop/remove/uninstall Kubeshark from the cluster. Do NOT use kubectl delete or helm uninstall directly - this tool ensures clean removal of all Kubeshark resources.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"release_namespace": {
						"type": "string",
						"description": "Namespace where Kubeshark is installed (default: 'default')"
					}
				}
			}`),
		},
		{
			Name:        "check_kubeshark_status",
			Description: "REQUIRED: Use this tool to check if Kubeshark is running/installed/deployed. Do NOT use kubectl get pods or other commands - this tool provides accurate status information and indicates whether other Kubeshark tools can be used.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"release_namespace": {
						"type": "string",
						"description": "Namespace where Kubeshark is installed (default: 'default')"
					}
				}
			}`),
		},
	}

	s.sendResult(req.ID, mcpListToolsResult{Tools: tools})
}

func (s *mcpServer) handleListPrompts(req *jsonRPCRequest) {
	prompts := []mcpPrompt{
		{
			Name:        "kubeshark_usage",
			Description: "Instructions for using Kubeshark MCP tools correctly",
		},
	}
	s.sendResult(req.ID, mcpListPromptsResult{Prompts: prompts})
}

func (s *mcpServer) handleGetPrompt(req *jsonRPCRequest) {
	var params mcpGetPromptParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	if params.Name != "kubeshark_usage" {
		s.sendError(req.ID, -32602, "Unknown prompt", params.Name)
		return
	}

	result := mcpGetPromptResult{
		Messages: []mcpPromptMessage{
			{
				Role: "user",
				Content: mcpContent{
					Type: "text",
					Text: `When working with Kubeshark, you MUST use the MCP tools provided - do NOT use kubectl, helm, or other CLI commands directly.

IMPORTANT RULES:
1. To check Kubeshark status: ALWAYS use 'check_kubeshark_status' tool (NOT 'kubectl get pods')
2. To start Kubeshark: ALWAYS use 'start_kubeshark' tool (NOT 'kubectl apply' or 'helm install')
3. To stop Kubeshark: ALWAYS use 'stop_kubeshark' tool (NOT 'kubectl delete' or 'helm uninstall')
4. To query captured traffic: Use 'list_workloads', 'list_api_calls', 'get_api_call', 'get_api_stats'

The MCP tools handle all the complexity of deployment, configuration, and API communication. Using kubectl/helm directly may cause issues or provide incomplete information.

When the user asks about Kubeshark status, traffic, or wants to start/stop Kubeshark, use the appropriate MCP tool immediately.`,
				},
			},
		},
	}
	s.sendResult(req.ID, result)
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
	case "start_kubeshark":
		result, isError = s.callStartKubeshark(params.Arguments)
	case "stop_kubeshark":
		result, isError = s.callStopKubeshark(params.Arguments)
	case "check_kubeshark_status":
		result, isError = s.callCheckKubesharkStatus(params.Arguments)
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
	if errMsg := s.ensureBackendConnection(); errMsg != "" {
		return errMsg, true
	}

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
	if errMsg := s.ensureBackendConnection(); errMsg != "" {
		return errMsg, true
	}

	query := url.Values{}
	for _, key := range []string{"src_ns", "src_pod", "dst_ns", "dst_pod", "dst_svc", "proto", "method", "path", "status", "group_by"} {
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
	if errMsg := s.ensureBackendConnection(); errMsg != "" {
		return errMsg, true
	}

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
	if errMsg := s.ensureBackendConnection(); errMsg != "" {
		return errMsg, true
	}

	query := url.Values{}
	for _, key := range []string{"ns", "svc", "pod", "group_by"} {
		if v, ok := args[key].(string); ok && v != "" {
			query.Set(key, v)
		}
	}

	return s.doGet("/stats", query)
}

func (s *mcpServer) callStartKubeshark(args map[string]any) (string, bool) {
	// Build the kubeshark tap command
	cmdArgs := []string{"tap"}

	// Add pod regex if provided
	if v, ok := args["pod_regex"].(string); ok && v != "" {
		cmdArgs = append(cmdArgs, v)
	}

	// Add namespaces if provided
	if v, ok := args["namespaces"].(string); ok && v != "" {
		namespaces := strings.Split(v, ",")
		for _, ns := range namespaces {
			ns = strings.TrimSpace(ns)
			if ns != "" {
				cmdArgs = append(cmdArgs, "-n", ns)
			}
		}
	}

	// Get release namespace
	releaseNamespace := config.Config.Tap.Release.Namespace
	if v, ok := args["release_namespace"].(string); ok && v != "" {
		releaseNamespace = v
		cmdArgs = append(cmdArgs, "-s", v)
	}

	// Add any custom --tap-set flags from MCP config
	for _, setFlag := range s.tapSetFlags {
		cmdArgs = append(cmdArgs, "--set", setFlag)
	}

	// Execute the command in headless mode (no browser popup)
	cmdArgs = append(cmdArgs, "--set", "headless=true")

	// Log progress to stderr (MCP clients can see this in their logs)
	logProgress := func(msg string) {
		_, _ = fmt.Fprintf(os.Stderr, "[kubeshark-mcp] %s\n", msg)
	}

	logProgress(fmt.Sprintf("Starting Kubeshark: %s %s", misc.Program, strings.Join(cmdArgs, " ")))

	// Start the command in the background (kubeshark tap runs continuously)
	cmd := exec.Command(misc.Program, cmdArgs...)
	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("Failed to start Kubeshark: %v", err), true
	}

	logProgress("Kubeshark process started, waiting for pods to be ready...")

	// Wait for Kubeshark to be ready (poll for pods)
	kubernetesProvider, err := getKubernetesProviderForCli(true, true)
	if err != nil {
		return fmt.Sprintf("Kubeshark command started but failed to check status: %v", err), true
	}

	// Poll for up to 2 minutes for pods to be ready
	ready := false
	for i := 0; i < 24; i++ { // 24 * 5s = 120s
		pods, err := kubernetesProvider.ListPodsByAppLabel(context.Background(), releaseNamespace, map[string]string{"app.kubernetes.io/name": "kubeshark"})
		if err == nil && len(pods) > 0 {
			// Check if at least hub pod is running
			for _, pod := range pods {
				if strings.Contains(pod.Name, "hub") && pod.Status.Phase == "Running" {
					ready = true
					logProgress("Hub pod is running")
					break
				}
			}
			if !ready {
				logProgress(fmt.Sprintf("Waiting for hub pod... (%d pods found, checking status)", len(pods)))
			}
		} else if i > 0 && i%3 == 0 {
			logProgress("Waiting for Kubeshark pods to be created...")
		}
		if ready {
			break
		}
		select {
		case <-context.Background().Done():
			return "Kubeshark start interrupted", true
		default:
			// Sleep 5 seconds before next check
			<-time.After(5 * time.Second)
		}
	}

	if !ready {
		logProgress("Timeout waiting for pods to be ready")
		return fmt.Sprintf("Kubeshark started but pods are not ready yet. Command: %s %s\nCheck status with check_kubeshark_status tool.", misc.Program, strings.Join(cmdArgs, " ")), false
	}

	// Reset backend connection state so next API call will re-establish connection
	s.backendMu.Lock()
	s.backendInitialized = false
	s.backendMu.Unlock()

	return fmt.Sprintf("Kubeshark started successfully and is ready.\nCommand: %s %s", misc.Program, strings.Join(cmdArgs, " ")), false
}

func (s *mcpServer) callStopKubeshark(args map[string]any) (string, bool) {
	// Build the kubeshark clean command
	cmdArgs := []string{"clean"}

	// Add release namespace if provided
	if v, ok := args["release_namespace"].(string); ok && v != "" {
		cmdArgs = append(cmdArgs, "-s", v)
	}

	cmd := exec.Command(misc.Program, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to stop Kubeshark: %v\nOutput: %s", err, string(output)), true
	}

	// Reset backend connection state
	s.backendMu.Lock()
	s.backendInitialized = false
	s.backendMu.Unlock()

	return fmt.Sprintf("Kubeshark stopped successfully.\nCommand: %s %s\nOutput: %s", misc.Program, strings.Join(cmdArgs, " "), string(output)), false
}

func (s *mcpServer) callCheckKubesharkStatus(args map[string]any) (string, bool) {
	namespace := config.Config.Tap.Release.Namespace
	if v, ok := args["release_namespace"].(string); ok && v != "" {
		namespace = v
	}

	kubernetesProvider, err := getKubernetesProviderForCli(true, true)
	if err != nil {
		return fmt.Sprintf("Failed to get Kubernetes provider: %v", err), true
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exists, err := kubernetesProvider.DoesServiceExist(ctx, namespace, kubernetes.FrontServiceName)
	if err != nil {
		return fmt.Sprintf("Error checking Kubeshark status: %v", err), true
	}

	if exists {
		return fmt.Sprintf(`Kubeshark is running in namespace '%s'.

Available tools:
- stop_kubeshark: Stop Kubeshark and remove resources
- list_workloads: List pods, services, namespaces with observed traffic
- list_api_calls: Query captured L7 API transactions (HTTP, gRPC, etc.)
- get_api_call: Get detailed info about a specific API call
- get_api_stats: Get aggregated API statistics`, namespace), false
	}

	return fmt.Sprintf(`Kubeshark is not running in namespace '%s'.

Available tools:
- start_kubeshark: Start Kubeshark to capture network traffic`, namespace), false
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
	defer func() { _ = resp.Body.Close() }()

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
	_, _ = fmt.Fprintln(s.stdout, string(data))
}
