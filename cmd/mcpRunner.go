package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Arguments   []mcpPromptArg    `json:"arguments,omitempty"`
}

type mcpPromptArg struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
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

// Hub MCP API response types

type hubMCPResponse struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	Tools       []hubMCPTool    `json:"tools"`
	Prompts     []hubMCPPrompt  `json:"prompts"`
}

type hubMCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type hubMCPPrompt struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Arguments   []hubMCPPromptArg   `json:"arguments,omitempty"`
}

type hubMCPPromptArg struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// MCP Server

type mcpServer struct {
	hubBaseURL         string
	httpClient         *http.Client
	stdin              io.Reader
	stdout             io.Writer
	backendInitialized bool
	backendMu          sync.Mutex
	setFlags           []string // --set flags to pass to 'kubeshark tap' when starting
	directURL          string   // If set, connect directly to this URL (no kubectl/proxy)
	urlMode            bool     // True when using direct URL mode
	allowDestructive   bool     // If true, enable start/stop tools
	cachedHubMCP       *hubMCPResponse // Cached tools/prompts from Hub
	cachedAt           time.Time       // When the cache was populated
	hubMCPMu           sync.Mutex
}

const hubMCPCacheTTL = 5 * time.Minute

func runMCPWithConfig(setFlags []string, directURL string, allowDestructive bool) {
	// Disable zerolog output to stderr (MCP uses stdio)
	zerolog.SetGlobalLevel(zerolog.Disabled)

	server := &mcpServer{
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		stdin:            os.Stdin,
		stdout:           os.Stdout,
		setFlags:         setFlags,
		directURL:        directURL,
		urlMode:          directURL != "",
		allowDestructive: allowDestructive,
	}

	// If URL mode, validate the URL is accessible on startup
	if server.urlMode {
		if err := server.validateDirectURL(); err != nil {
			fmt.Fprintf(os.Stderr, "[kubeshark-mcp] Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "[kubeshark-mcp] Connected to Kubeshark at %s\n", directURL)
	}

	server.run()
}

// validateDirectURL checks that the direct URL is accessible
func (s *mcpServer) validateDirectURL() error {
	// Normalize URL - ensure it doesn't end with /
	urlStr := strings.TrimSuffix(s.directURL, "/")
	s.directURL = urlStr

	// Use a short timeout for validation
	client := &http.Client{Timeout: 10 * time.Second}

	// Try to reach the MCP API base endpoint which returns tool definitions
	testURL := fmt.Sprintf("%s/api/mcp", urlStr)
	resp, err := client.Get(testURL)
	if err != nil {
		return fmt.Errorf("cannot connect to Kubeshark at %s: %v", urlStr, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Try to parse the MCP response to validate it's a valid Kubeshark endpoint
	var mcpInfo hubMCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&mcpInfo); err != nil {
		return fmt.Errorf("invalid response from Kubeshark at %s: %v", urlStr, err)
	}

	// Verify it looks like a valid MCP response
	if mcpInfo.Name == "" && len(mcpInfo.Tools) == 0 {
		return fmt.Errorf("kubeshark at %s does not appear to have MCP enabled", urlStr)
	}

	// Set the hub base URL
	s.hubBaseURL = fmt.Sprintf("%s/api/mcp", urlStr)
	s.backendInitialized = true
	return nil
}

// ensureBackendConnection establishes connection to Kubeshark backend if not already connected
// Returns an error message if connection fails, empty string on success
func (s *mcpServer) ensureBackendConnection() string {
	s.backendMu.Lock()
	defer s.backendMu.Unlock()

	if s.backendInitialized {
		return ""
	}

	// In URL mode, connection was validated at startup
	if s.urlMode {
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

	// Start proxy to frontend and verify connectivity
	frontURL := kubernetes.GetProxyOnPort(config.Config.Tap.Proxy.Front.Port)
	response, err := http.Get(fmt.Sprintf("%s/", frontURL))
	if response != nil && response.Body != nil {
		defer func() { _ = response.Body.Close() }()
	}
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

// fetchHubMCP fetches tools and prompts from the Hub's /api/mcp endpoint
// Returns nil if Hub is not available or returns an error
func (s *mcpServer) fetchHubMCP() *hubMCPResponse {
	s.hubMCPMu.Lock()
	defer s.hubMCPMu.Unlock()

	// Return cached if available and not expired
	if s.cachedHubMCP != nil && time.Since(s.cachedAt) < hubMCPCacheTTL {
		return s.cachedHubMCP
	}

	// Ensure backend connection first
	if errMsg := s.ensureBackendConnection(); errMsg != "" {
		return nil
	}

	// Fetch from Hub - the base URL is like http://host/api/mcp, we need http://host/api/mcp
	// The Hub's MCP info endpoint is at the base path
	resp, err := s.httpClient.Get(s.hubBaseURL)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil
	}

	var hubMCP hubMCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&hubMCP); err != nil {
		return nil
	}

	s.cachedHubMCP = &hubMCP
	s.cachedAt = time.Now()
	return s.cachedHubMCP
}

// invalidateHubMCPCache clears the cached Hub MCP data
func (s *mcpServer) invalidateHubMCPCache() {
	s.hubMCPMu.Lock()
	defer s.hubMCPMu.Unlock()
	s.cachedHubMCP = nil
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

	// Check for scanner errors (e.g., stdin closed, read errors)
	if err := scanner.Err(); err != nil {
		writeErrorToStderr("[kubeshark-mcp] Scanner error: %v", err)
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
	var instructions string
	if s.urlMode {
		instructions = fmt.Sprintf(`Kubeshark MCP Server - Connected to: %s

This server provides read-only access to Kubeshark's traffic analysis capabilities.
Cluster management tools (start/stop) are disabled in URL mode.

Available tools for traffic analysis:
- list_workloads: List pods, services, namespaces with observed traffic
- list_api_calls: Query L7 API transactions (HTTP, gRPC, etc.)
- list_l4_flows: View L4 (TCP/UDP) network flows
- get_api_stats: Get aggregated API statistics
- And more - use tools/list to see all available tools

Use the MCP tools directly - do NOT use kubectl or curl to access Kubeshark.`, s.directURL)
	} else if s.allowDestructive {
		instructions = `Kubeshark MCP Server - Proxy Mode (Destructive Operations ENABLED)

This server proxies to a Kubeshark deployment in your Kubernetes cluster.

⚠️ DESTRUCTIVE OPERATIONS ENABLED (--allow-destructive flag is set):
- start_kubeshark: Deploys Kubeshark to your cluster (runs 'kubeshark tap')
- stop_kubeshark: Removes Kubeshark from your cluster (runs 'kubeshark clean')

ALWAYS confirm with the user before using start_kubeshark or stop_kubeshark.

Safe operations:
- check_kubeshark_status: Check if Kubeshark is running (read-only)

Traffic analysis tools (require Kubeshark to be running):
- list_workloads, list_api_calls, list_l4_flows, get_api_stats, and more

Use the MCP tools - do NOT use kubectl, helm, or curl directly.`
	} else {
		instructions = `Kubeshark MCP Server - Proxy Mode (Read-Only)

This server proxies to an existing Kubeshark deployment in your Kubernetes cluster.

Destructive operations (start/stop) are DISABLED for safety.
To enable them, restart with --allow-destructive flag.

Available operations:
- check_kubeshark_status: Check if Kubeshark is running (read-only)

Traffic analysis tools (require Kubeshark to be running):
- list_workloads, list_api_calls, list_l4_flows, get_api_stats, and more

Use the MCP tools - do NOT use kubectl, helm, or curl directly.`
	}

	result := mcpInitializeResult{
		ProtocolVersion: "2024-11-05",
		Instructions:    instructions,
	}
	result.ServerInfo.Name = "kubeshark-mcp"
	result.ServerInfo.Version = "1.0.0"

	s.sendResult(req.ID, result)
}

func (s *mcpServer) handleListTools(req *jsonRPCRequest) {
	var tools []mcpTool

	// Add check_kubeshark_status - safe, read-only operation that works in both modes
	tools = append(tools, mcpTool{
		Name:        "check_kubeshark_status",
		Description: "Safe: Checks if Kubeshark is currently running and accessible. In URL mode, confirms connectivity to the remote instance. In local mode, checks cluster pods. This is a read-only operation.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"release_namespace": {
					"type": "string",
					"description": "Namespace where Kubeshark is installed (default: 'default'). Only used in local mode."
				}
			}
		}`),
	})

	// Add destructive tools only if --allow-destructive flag was set (and not in URL mode)
	if !s.urlMode && s.allowDestructive {
		tools = append(tools, mcpTool{
			Name:        "start_kubeshark",
			Description: "⚠️ DESTRUCTIVE: Deploys Kubeshark to the Kubernetes cluster by running 'kubeshark tap'. This will create pods, services, and other resources in the cluster. ALWAYS confirm with the user before using this tool. Use check_kubeshark_status first to see if Kubeshark is already running.",
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
		})
		tools = append(tools, mcpTool{
			Name:        "stop_kubeshark",
			Description: "⚠️ DESTRUCTIVE: Removes Kubeshark from the Kubernetes cluster by running 'kubeshark clean'. This will delete all Kubeshark pods, services, and resources. All captured traffic data will be lost. ALWAYS confirm with the user before using this tool.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"release_namespace": {
						"type": "string",
						"description": "Namespace where Kubeshark is installed (default: 'default')"
					}
				}
			}`),
		})
	}

	// Fetch tools from Hub and merge
	if hubMCP := s.fetchHubMCP(); hubMCP != nil {
		for _, hubTool := range hubMCP.Tools {
			tools = append(tools, mcpTool(hubTool))
		}
	}

	s.sendResult(req.ID, mcpListToolsResult{Tools: tools})
}

func (s *mcpServer) handleListPrompts(req *jsonRPCRequest) {
	var prompts []mcpPrompt

	// Add local prompts
	prompts = append(prompts, mcpPrompt{
		Name:        "kubeshark_usage",
		Description: "Instructions for using Kubeshark MCP tools correctly",
	})

	// Fetch prompts from Hub and merge
	if hubMCP := s.fetchHubMCP(); hubMCP != nil {
		for _, hubPrompt := range hubMCP.Prompts {
			var args []mcpPromptArg
			for _, hubArg := range hubPrompt.Arguments {
				args = append(args, mcpPromptArg(hubArg))
			}
			prompts = append(prompts, mcpPrompt{
				Name:        hubPrompt.Name,
				Description: hubPrompt.Description,
				Arguments:   args,
			})
		}
	}

	s.sendResult(req.ID, mcpListPromptsResult{Prompts: prompts})
}

func (s *mcpServer) handleGetPrompt(req *jsonRPCRequest) {
	var params mcpGetPromptParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	// Handle local prompt
	if params.Name == "kubeshark_usage" {
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
4. To query captured traffic: Use the available traffic analysis tools

The MCP tools handle all the complexity of deployment, configuration, and API communication. Using kubectl/helm directly may cause issues or provide incomplete information.

When the user asks about Kubeshark status, traffic, or wants to start/stop Kubeshark, use the appropriate MCP tool immediately.`,
					},
				},
			},
		}
		s.sendResult(req.ID, result)
		return
	}

	// Check if it's a Hub prompt
	hubMCP := s.fetchHubMCP()
	if hubMCP != nil {
		for _, hubPrompt := range hubMCP.Prompts {
			if hubPrompt.Name == params.Name {
				// Generate prompt message from Hub prompt definition
				promptText := fmt.Sprintf("Task: %s\n\n%s", hubPrompt.Name, hubPrompt.Description)
				if len(hubPrompt.Arguments) > 0 {
					promptText += "\n\nParameters:\n"
					for _, arg := range hubPrompt.Arguments {
						required := ""
						if arg.Required {
							required = " (required)"
						}
						promptText += fmt.Sprintf("- %s%s: %s\n", arg.Name, required, arg.Description)
					}
				}
				promptText += "\n\nUse the appropriate Kubeshark MCP tools to complete this task."

				result := mcpGetPromptResult{
					Messages: []mcpPromptMessage{
						{
							Role: "user",
							Content: mcpContent{
								Type: "text",
								Text: promptText,
							},
						},
					},
				}
				s.sendResult(req.ID, result)
				return
			}
		}
	}

	s.sendError(req.ID, -32602, "Unknown prompt", params.Name)
}

func (s *mcpServer) handleCallTool(req *jsonRPCRequest) {
	var params mcpCallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	var result string
	var isError bool

	// Handle local CLI tools
	switch params.Name {
	case "start_kubeshark":
		if s.urlMode {
			result, isError = "This tool is not available in URL mode. Kubeshark is managed externally.", true
		} else if !s.allowDestructive {
			result, isError = "This tool requires --allow-destructive flag. Destructive operations are disabled for safety.", true
		} else {
			result, isError = s.callStartKubeshark(params.Arguments)
		}
		s.sendResult(req.ID, mcpCallToolResult{
			Content: []mcpContent{{Type: "text", Text: result}},
			IsError: isError,
		})
		return
	case "stop_kubeshark":
		if s.urlMode {
			result, isError = "This tool is not available in URL mode. Kubeshark is managed externally.", true
		} else if !s.allowDestructive {
			result, isError = "This tool requires --allow-destructive flag. Destructive operations are disabled for safety.", true
		} else {
			result, isError = s.callStopKubeshark(params.Arguments)
		}
		s.sendResult(req.ID, mcpCallToolResult{
			Content: []mcpContent{{Type: "text", Text: result}},
			IsError: isError,
		})
		return
	case "check_kubeshark_status":
		if s.urlMode {
			result, isError = fmt.Sprintf("Kubeshark is accessible at %s (URL mode - externally managed)", s.directURL), false
		} else {
			result, isError = s.callCheckKubesharkStatus(params.Arguments)
		}
		s.sendResult(req.ID, mcpCallToolResult{
			Content: []mcpContent{{Type: "text", Text: result}},
			IsError: isError,
		})
		return
	}

	// Forward Hub tools to the API
	result, isError = s.callHubTool(params.Name, params.Arguments)
	s.sendResult(req.ID, mcpCallToolResult{
		Content: []mcpContent{{Type: "text", Text: result}},
		IsError: isError,
	})
}

// callHubTool forwards a tool call to the Hub's MCP API
func (s *mcpServer) callHubTool(toolName string, args map[string]any) (string, bool) {
	if errMsg := s.ensureBackendConnection(); errMsg != "" {
		return errMsg, true
	}

	// Build the request body
	requestBody := map[string]any{
		"name":      toolName,
		"arguments": args,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Sprintf("Error encoding request: %v", err), true
	}

	// POST to /api/mcp/tools/call
	reqURL := s.hubBaseURL + "/tools/call"
	resp, err := s.httpClient.Post(reqURL, "application/json", bytes.NewReader(bodyBytes))
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

	// Add any custom --set flags from MCP config
	for _, setFlag := range s.setFlags {
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

	// Wait for the process in a goroutine to prevent zombie processes
	go func() {
		_ = cmd.Wait()
	}()

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
		// Sleep 5 seconds before next check
		time.Sleep(5 * time.Second)
	}

	if !ready {
		logProgress("Timeout waiting for pods to be ready")
		return fmt.Sprintf("Kubeshark started but pods are not ready yet. Command: %s %s\nCheck status with check_kubeshark_status tool.", misc.Program, strings.Join(cmdArgs, " ")), false
	}

	// Reset backend connection state so next API call will re-establish connection
	s.backendMu.Lock()
	s.backendInitialized = false
	s.backendMu.Unlock()

	// Invalidate cached tools/prompts so they're fetched from the new Hub
	s.invalidateHubMCPCache()

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

	// Invalidate cached tools/prompts
	s.invalidateHubMCPCache()

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

// listMCPTools prints available MCP tools to stdout
func listMCPTools(directURL string) {
	fmt.Println("MCP Tools")
	fmt.Println("=========")
	fmt.Println()

	// URL mode - no cluster management, connect directly
	if directURL != "" {
		fmt.Printf("URL Mode: %s\n\n", directURL)
		fmt.Println("Cluster management tools disabled (Kubeshark managed externally)")
		fmt.Println()

		hubURL := strings.TrimSuffix(directURL, "/") + "/api/mcp"
		fetchAndDisplayTools(hubURL, 30*time.Second)
		return
	}

	// Normal mode - show cluster management tools
	fmt.Println("Cluster Management:")
	fmt.Println("  check_kubeshark_status  Check if Kubeshark is running in the cluster")
	fmt.Println("  start_kubeshark         Start Kubeshark to capture traffic")
	fmt.Println("  stop_kubeshark          Stop Kubeshark and clean up resources")
	fmt.Println()

	// Establish proxy connection to Kubeshark
	fmt.Println("Connecting to Kubeshark...")
	hubURL, err := establishProxyConnection(30 * time.Second)
	if err != nil {
		fmt.Printf("\nKubeshark API: %v\n", err)
		return
	}

	fmt.Printf("Connected to: %s\n\n", hubURL)
	fetchAndDisplayTools(hubURL, 30*time.Second)
}

// establishProxyConnection sets up proxy to Kubeshark and returns the hub URL
func establishProxyConnection(timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	kubernetesProvider, err := getKubernetesProviderForCli(true, true)
	if err != nil {
		return "", fmt.Errorf("failed to get Kubernetes provider: %v", err)
	}

	// Check if Kubeshark services exist
	exists, err := kubernetesProvider.DoesServiceExist(ctx, config.Config.Tap.Release.Namespace, kubernetes.FrontServiceName)
	if err != nil {
		return "", fmt.Errorf("error checking Kubeshark status: %v", err)
	}
	if !exists {
		return "", fmt.Errorf("not running (use start_kubeshark to start)")
	}

	// Start proxy to frontend and verify connectivity
	frontURL := kubernetes.GetProxyOnPort(config.Config.Tap.Proxy.Front.Port)
	response, err := http.Get(fmt.Sprintf("%s/", frontURL))
	if response != nil && response.Body != nil {
		defer func() { _ = response.Body.Close() }()
	}
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
			return "", fmt.Errorf("couldn't connect to Kubeshark frontend")
		}
	}

	return fmt.Sprintf("%s/api/mcp", frontURL), nil
}

// fetchAndDisplayTools fetches tools from the Kubeshark API and displays them
func fetchAndDisplayTools(hubURL string, timeout time.Duration) {
	client := &http.Client{Timeout: timeout}

	// Fetch tools list from /api/mcp endpoint
	resp, err := client.Get(strings.TrimSuffix(hubURL, "/mcp") + "/mcp")
	if err != nil {
		fmt.Printf("Kubeshark API: Connection failed (%v)\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Parse the response using the Hub MCP response format
	var mcpInfo hubMCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&mcpInfo); err != nil {
		fmt.Printf("Kubeshark API: Connected (couldn't parse response: %v)\n", err)
		return
	}

	if len(mcpInfo.Tools) == 0 {
		fmt.Println("Kubeshark API: Connected (no tools available)")
		return
	}

	fmt.Println("Traffic Analysis Tools:")
	for _, tool := range mcpInfo.Tools {
		desc := tool.Description
		if len(desc) > 55 {
			desc = desc[:52] + "..."
		}
		fmt.Printf("  %-24s %s\n", tool.Name, desc)
	}

	if len(mcpInfo.Prompts) > 0 {
		fmt.Println()
		fmt.Println("Prompts:")
		for _, prompt := range mcpInfo.Prompts {
			desc := prompt.Description
			if len(desc) > 55 {
				desc = desc[:52] + "..."
			}
			fmt.Printf("  %-24s %s\n", prompt.Name, desc)
		}
	}
}

// printMCPConfig outputs the Claude Desktop configuration JSON
func printMCPConfig(directURL string, kubeconfig string) {
	// Get the path to the kubeshark binary
	binaryPath, err := os.Executable()
	if err != nil {
		binaryPath = "kubeshark"
	}

	// Build args
	args := []string{"mcp"}
	if directURL != "" {
		args = append(args, "--url", directURL)
	} else if kubeconfig != "" {
		args = append(args, "--kubeconfig", kubeconfig)
	} else {
		// Default to user's kubeconfig
		kubeconfig = config.Config.KubeConfigPath()
		if kubeconfig != "" {
			args = append(args, "--kubeconfig", kubeconfig)
		}
	}

	// Build config structure
	mcpConfig := map[string]any{
		"mcpServers": map[string]any{
			"kubeshark": map[string]any{
				"command": binaryPath,
				"args":    args,
			},
		},
	}

	// Output as JSON
	output, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating config: %v\n", err)
		return
	}
	fmt.Println(string(output))
}
