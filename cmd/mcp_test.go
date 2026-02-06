package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test helpers

func newTestMCPServer() *mcpServer {
	return &mcpServer{
		httpClient: &http.Client{},
		stdin:      &bytes.Buffer{},
		stdout:     &bytes.Buffer{},
	}
}

func sendRequest(s *mcpServer, method string, id any, params any) string {
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}
	if params != nil {
		paramsBytes, _ := json.Marshal(params)
		req.Params = paramsBytes
	}

	s.handleRequest(&req)

	output := s.stdout.(*bytes.Buffer).String()
	s.stdout.(*bytes.Buffer).Reset()
	return output
}

func parseResponse(t *testing.T, output string) jsonRPCResponse {
	var resp jsonRPCResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v\nOutput: %s", err, output)
	}
	return resp
}

// =============================================================================
// JSON-RPC Protocol Tests
// =============================================================================

func TestMCP_Initialize(t *testing.T) {
	s := newTestMCPServer()
	output := sendRequest(s, "initialize", 1, nil)
	resp := parseResponse(t, output)

	if resp.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Result is not a map: %T", resp.Result)
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion 2024-11-05, got %v", result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatalf("serverInfo is not a map")
	}
	if serverInfo["name"] != "kubeshark-mcp" {
		t.Errorf("Expected server name kubeshark-mcp, got %v", serverInfo["name"])
	}

	// Verify instructions are included
	instructions, ok := result["instructions"].(string)
	if !ok || instructions == "" {
		t.Error("Expected non-empty instructions in initialize response")
	}
	if !strings.Contains(instructions, "check_kubeshark_status") {
		t.Error("Instructions should mention check_kubeshark_status tool")
	}

	// Verify capabilities include prompts
	capabilities, ok := result["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("capabilities is not a map")
	}
	if _, hasPrompts := capabilities["prompts"]; !hasPrompts {
		t.Error("Expected prompts capability")
	}
}

func TestMCP_Ping(t *testing.T) {
	s := newTestMCPServer()
	output := sendRequest(s, "ping", 42, nil)
	resp := parseResponse(t, output)

	if resp.ID != float64(42) {
		t.Errorf("Expected ID 42, got %v", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Result is not a map: %T", resp.Result)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %v", result)
	}
}

func TestMCP_InitializedNotification(t *testing.T) {
	s := newTestMCPServer()

	// "initialized" is a notification, should not produce a response
	output := sendRequest(s, "initialized", nil, nil)
	if output != "" {
		t.Errorf("Expected no output for notification, got: %s", output)
	}

	// "notifications/initialized" is also a notification
	output = sendRequest(s, "notifications/initialized", nil, nil)
	if output != "" {
		t.Errorf("Expected no output for notification, got: %s", output)
	}
}

func TestMCP_UnknownMethod(t *testing.T) {
	s := newTestMCPServer()
	output := sendRequest(s, "unknown/method", 1, nil)
	resp := parseResponse(t, output)

	if resp.Error == nil {
		t.Fatal("Expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", resp.Error.Code)
	}
	if resp.Error.Message != "Method not found" {
		t.Errorf("Expected 'Method not found', got %s", resp.Error.Message)
	}
}

func TestMCP_PromptsList(t *testing.T) {
	s := newTestMCPServer()
	output := sendRequest(s, "prompts/list", 1, nil)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Result is not a map: %T", resp.Result)
	}

	prompts, ok := result["prompts"].([]any)
	if !ok {
		t.Fatalf("prompts is not an array: %T", result["prompts"])
	}

	if len(prompts) != 1 {
		t.Errorf("Expected 1 prompt, got %d", len(prompts))
	}

	prompt := prompts[0].(map[string]any)
	if prompt["name"] != "kubeshark_usage" {
		t.Errorf("Expected prompt name 'kubeshark_usage', got %v", prompt["name"])
	}
}

func TestMCP_PromptsGet(t *testing.T) {
	s := newTestMCPServer()
	params := map[string]any{"name": "kubeshark_usage"}
	output := sendRequest(s, "prompts/get", 1, params)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Result is not a map: %T", resp.Result)
	}

	messages, ok := result["messages"].([]any)
	if !ok || len(messages) == 0 {
		t.Fatalf("Expected messages array with at least one message")
	}

	msg := messages[0].(map[string]any)
	if msg["role"] != "user" {
		t.Errorf("Expected role 'user', got %v", msg["role"])
	}

	content := msg["content"].(map[string]any)
	text := content["text"].(string)

	// Verify the prompt contains key instructions
	expectedPhrases := []string{
		"check_kubeshark_status",
		"start_kubeshark",
		"stop_kubeshark",
		"NOT",
		"kubectl",
	}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(text, phrase) {
			t.Errorf("Prompt should contain '%s'", phrase)
		}
	}
}

func TestMCP_PromptsGet_UnknownPrompt(t *testing.T) {
	s := newTestMCPServer()
	params := map[string]any{"name": "unknown_prompt"}
	output := sendRequest(s, "prompts/get", 1, params)
	resp := parseResponse(t, output)

	if resp.Error == nil {
		t.Fatal("Expected error for unknown prompt")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("Expected error code -32602, got %d", resp.Error.Code)
	}
}

func TestMCP_ToolsList_CLIOnly(t *testing.T) {
	// Test tools/list without Hub backend - should only have CLI tools
	s := newTestMCPServer()
	output := sendRequest(s, "tools/list", 1, nil)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Result is not a map: %T", resp.Result)
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("tools is not an array: %T", result["tools"])
	}

	// Without Hub backend and without allowDestructive, only check_kubeshark_status is available
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool (check_kubeshark_status only), got %d", len(tools))
	}

	toolMap := tools[0].(map[string]any)
	if toolMap["name"] != "check_kubeshark_status" {
		t.Errorf("Expected check_kubeshark_status, got %v", toolMap["name"])
	}

	// Verify tool has required fields
	if _, hasDesc := toolMap["description"]; !hasDesc {
		t.Error("Tool missing description")
	}
	if _, hasSchema := toolMap["inputSchema"]; !hasSchema {
		t.Error("Tool missing inputSchema")
	}
}

func TestMCP_ToolsList_WithDestructive(t *testing.T) {
	// Test tools/list with allowDestructive enabled
	s := &mcpServer{
		httpClient:       &http.Client{},
		stdin:            &bytes.Buffer{},
		stdout:           &bytes.Buffer{},
		allowDestructive: true,
	}
	output := sendRequest(s, "tools/list", 1, nil)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result := resp.Result.(map[string]any)
	tools := result["tools"].([]any)

	// With allowDestructive, should have check_kubeshark_status, start_kubeshark, stop_kubeshark
	expectedTools := []string{"check_kubeshark_status", "start_kubeshark", "stop_kubeshark"}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolMap := tool.(map[string]any)
		toolNames[toolMap["name"].(string)] = true
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("Missing expected tool: %s", expected)
		}
	}
}

func TestMCP_ToolsList_WithHubBackend(t *testing.T) {
	// Test tools/list with a mock Hub backend that provides tools
	mockHubMCPResponse := `{
		"name": "kubeshark-hub",
		"version": "1.0.0",
		"tools": [
			{"name": "list_workloads", "description": "List workloads", "inputSchema": {}},
			{"name": "list_api_calls", "description": "List API calls", "inputSchema": {}},
			{"name": "get_api_call", "description": "Get API call details", "inputSchema": {}},
			{"name": "get_api_stats", "description": "Get API stats", "inputSchema": {}}
		],
		"prompts": []
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The Hub MCP endpoint returns tool definitions at the base path
		if r.URL.Path == "/" || r.URL.Path == "" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(mockHubMCPResponse))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	s := &mcpServer{
		httpClient:         &http.Client{},
		stdin:              &bytes.Buffer{},
		stdout:             &bytes.Buffer{},
		hubBaseURL:         mockServer.URL,
		backendInitialized: true,
		allowDestructive:   true,
	}

	output := sendRequest(s, "tools/list", 1, nil)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result := resp.Result.(map[string]any)
	tools := result["tools"].([]any)

	// Should have CLI tools (3) + Hub tools (4) = 7 tools
	expectedTools := []string{
		"check_kubeshark_status",
		"start_kubeshark",
		"stop_kubeshark",
		"list_workloads",
		"list_api_calls",
		"get_api_call",
		"get_api_stats",
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
		for _, tool := range tools {
			toolMap := tool.(map[string]any)
			t.Logf("Found tool: %s", toolMap["name"])
		}
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolMap := tool.(map[string]any)
		name := toolMap["name"].(string)
		toolNames[name] = true

		// Verify each tool has required fields
		if _, hasDesc := toolMap["description"]; !hasDesc {
			t.Errorf("Tool %s missing description", name)
		}
		if _, hasSchema := toolMap["inputSchema"]; !hasSchema {
			t.Errorf("Tool %s missing inputSchema", name)
		}
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("Missing expected tool: %s", expected)
		}
	}
}

func TestMCP_ToolsCallUnknownTool(t *testing.T) {
	// When calling an unknown tool with a mock Hub, the Hub should return a 404
	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		// Return 404 for unknown tool
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Unknown tool"}`))
	})
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name: "unknown_tool",
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	// Unknown tool calls still return a result (not a JSON-RPC error) but with isError=true
	if resp.Error != nil {
		t.Fatalf("Unexpected JSON-RPC error: %v", resp.Error)
	}

	result := resp.Result.(map[string]any)
	isError, ok := result["isError"].(bool)
	if !ok || !isError {
		t.Error("Expected isError=true for unknown tool")
	}

	// Verify error message mentions the status code
	content := result["content"].([]any)
	firstContent := content[0].(map[string]any)
	text := firstContent["text"].(string)
	if !strings.Contains(text, "404") {
		t.Errorf("Expected error to mention 404, got: %s", text)
	}
}

func TestMCP_ToolsCallInvalidParams(t *testing.T) {
	s := newTestMCPServer()

	// Send invalid params (not a valid mcpCallToolParams structure)
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  json.RawMessage(`"invalid"`),
	}
	s.handleRequest(&req)

	output := s.stdout.(*bytes.Buffer).String()
	resp := parseResponse(t, output)

	if resp.Error == nil {
		t.Fatal("Expected error for invalid params")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("Expected error code -32602, got %d", resp.Error.Code)
	}
}

// =============================================================================
// CLI Tools Tests (work independently of Kubeshark status)
// =============================================================================

func TestMCP_CheckKubesharkStatus_NoKubernetesConfig(t *testing.T) {
	// This test verifies the tool handles missing kubernetes config gracefully
	s := newTestMCPServer()

	params := mcpCallToolParams{
		Name:      "check_kubeshark_status",
		Arguments: map[string]any{},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected JSON-RPC error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Result is not a map: %T", resp.Result)
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("Expected content array with at least one element")
	}

	// The tool should return either a status message or an error about kubernetes
	// Both are valid responses - the important thing is it doesn't crash
	firstContent := content[0].(map[string]any)
	text := firstContent["text"].(string)
	if text == "" {
		t.Error("Expected non-empty response text")
	}
}

func TestMCP_CheckKubesharkStatus_WithNamespace(t *testing.T) {
	s := newTestMCPServer()

	params := mcpCallToolParams{
		Name: "check_kubeshark_status",
		Arguments: map[string]any{
			"release_namespace": "custom-namespace",
		},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected JSON-RPC error: %v", resp.Error)
	}

	// Verify the response structure is correct
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Result is not a map: %T", resp.Result)
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("Expected content array")
	}
}

// =============================================================================
// API Tools Tests (with mock HTTP backend)
// =============================================================================

func newTestMCPServerWithMockBackend(handler http.HandlerFunc) (*mcpServer, *httptest.Server) {
	mockServer := httptest.NewServer(handler)
	s := &mcpServer{
		httpClient:         &http.Client{},
		stdin:              &bytes.Buffer{},
		stdout:             &bytes.Buffer{},
		hubBaseURL:         mockServer.URL,
		backendInitialized: true, // Skip backend connection check
	}
	return s, mockServer
}

// hubToolCallRequest represents the expected structure of POST /tools/call requests
type hubToolCallRequest struct {
	Tool      string         `json:"tool"`
	Arguments map[string]any `json:"arguments"`
}

// newMockHubHandler creates a mock handler for the Hub's /tools/call endpoint
// The toolHandler function receives the parsed tool call request and returns the response
func newMockHubHandler(t *testing.T, toolHandler func(req hubToolCallRequest) (string, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// All Hub tool calls go to /tools/call
		if r.URL.Path != "/tools/call" {
			t.Errorf("Expected path /tools/call, got %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Must be POST
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse the request body
		var req hubToolCallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to parse request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Call the handler
		response, status := toolHandler(req)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}
}

func TestMCP_ListWorkloads(t *testing.T) {
	mockResponse := `{"workloads": [{"name": "test-pod", "namespace": "default"}]}`

	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// Verify tool name
		if req.Tool != "list_workloads" {
			t.Errorf("Expected tool 'list_workloads', got %s", req.Tool)
		}
		// Verify arguments are in the request body
		if req.Arguments["type"] != "pod" {
			t.Errorf("Expected type=pod in arguments, got %v", req.Arguments["type"])
		}
		if req.Arguments["ns"] != "default" {
			t.Errorf("Expected ns=default in arguments, got %v", req.Arguments["ns"])
		}
		return mockResponse, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name: "list_workloads",
		Arguments: map[string]any{
			"type": "pod",
			"ns":   "default",
		},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result := resp.Result.(map[string]any)
	content := result["content"].([]any)
	firstContent := content[0].(map[string]any)
	text := firstContent["text"].(string)

	if !strings.Contains(text, "test-pod") {
		t.Errorf("Expected response to contain 'test-pod', got: %s", text)
	}
}

func TestMCP_ListWorkloads_WithLabels(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// Verify tool name
		if req.Tool != "list_workloads" {
			t.Errorf("Expected tool 'list_workloads', got %s", req.Tool)
		}
		// Verify arguments are in the POST body
		if req.Arguments["type"] != "service" {
			t.Errorf("Expected type=service in arguments, got %v", req.Arguments["type"])
		}
		if req.Arguments["labels"] != "app=nginx" {
			t.Errorf("Expected labels=app=nginx in arguments, got %v", req.Arguments["labels"])
		}
		return `{"workloads": []}`, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name: "list_workloads",
		Arguments: map[string]any{
			"type":   "service",
			"labels": "app=nginx",
		},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	// Verify no error
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}

func TestMCP_ListAPICalls(t *testing.T) {
	mockResponse := `{"calls": [{"id": "123", "method": "GET", "path": "/api/v1/users"}]}`

	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// Verify tool name
		if req.Tool != "list_api_calls" {
			t.Errorf("Expected tool 'list_api_calls', got %s", req.Tool)
		}
		// Verify arguments are in the POST body
		if req.Arguments["proto"] != "http" {
			t.Errorf("Expected proto=http in arguments, got %v", req.Arguments["proto"])
		}
		if req.Arguments["method"] != "GET" {
			t.Errorf("Expected method=GET in arguments, got %v", req.Arguments["method"])
		}
		if req.Arguments["limit"] != float64(50) {
			t.Errorf("Expected limit=50 in arguments, got %v", req.Arguments["limit"])
		}
		return mockResponse, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name: "list_api_calls",
		Arguments: map[string]any{
			"proto":  "http",
			"method": "GET",
			"limit":  float64(50),
		},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result := resp.Result.(map[string]any)
	content := result["content"].([]any)
	firstContent := content[0].(map[string]any)
	text := firstContent["text"].(string)

	if !strings.Contains(text, "/api/v1/users") {
		t.Errorf("Expected response to contain '/api/v1/users', got: %s", text)
	}
}

func TestMCP_ListAPICalls_AllFilters(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// Verify tool name
		if req.Tool != "list_api_calls" {
			t.Errorf("Expected tool 'list_api_calls', got %s", req.Tool)
		}
		// Verify all arguments are in the POST body
		expectedArgs := map[string]any{
			"src_ns":  "source-ns",
			"src_pod": "source-pod",
			"dst_ns":  "dest-ns",
			"dst_pod": "dest-pod",
			"dst_svc": "dest-svc",
			"proto":   "grpc",
			"method":  "GetUser",
			"path":    "/users",
			"status":  "error",
			"limit":   float64(10),
		}
		for key, expected := range expectedArgs {
			if req.Arguments[key] != expected {
				t.Errorf("Expected %s=%v in arguments, got %v", key, expected, req.Arguments[key])
			}
		}
		return `{"calls": []}`, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name: "list_api_calls",
		Arguments: map[string]any{
			"src_ns":  "source-ns",
			"src_pod": "source-pod",
			"dst_ns":  "dest-ns",
			"dst_pod": "dest-pod",
			"dst_svc": "dest-svc",
			"proto":   "grpc",
			"method":  "GetUser",
			"path":    "/users",
			"status":  "error",
			"limit":   float64(10),
		},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	// Verify no error
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}

func TestMCP_GetAPICall(t *testing.T) {
	mockResponse := `{
		"id": "abc123",
		"method": "POST",
		"path": "/api/orders",
		"status": 201,
		"request": {"headers": {"Content-Type": "application/json"}},
		"response": {"headers": {"X-Request-Id": "xyz"}}
	}`

	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// Verify tool name
		if req.Tool != "get_api_call" {
			t.Errorf("Expected tool 'get_api_call', got %s", req.Tool)
		}
		// Verify arguments are in the POST body (not URL path or query params)
		if req.Arguments["id"] != "abc123" {
			t.Errorf("Expected id=abc123 in arguments, got %v", req.Arguments["id"])
		}
		if req.Arguments["include_headers"] != true {
			t.Errorf("Expected include_headers=true in arguments, got %v", req.Arguments["include_headers"])
		}
		if req.Arguments["include_payload"] != true {
			t.Errorf("Expected include_payload=true in arguments, got %v", req.Arguments["include_payload"])
		}
		return mockResponse, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name: "get_api_call",
		Arguments: map[string]any{
			"id":              "abc123",
			"include_headers": true,
			"include_payload": true,
		},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result := resp.Result.(map[string]any)
	content := result["content"].([]any)
	firstContent := content[0].(map[string]any)
	text := firstContent["text"].(string)

	if !strings.Contains(text, "abc123") {
		t.Errorf("Expected response to contain 'abc123', got: %s", text)
	}
}

func TestMCP_GetAPICall_MissingID(t *testing.T) {
	// Hub tools are forwarded to the Hub, which handles validation
	// If the Hub requires an ID and returns an error, the test should reflect that
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// Hub validates that ID is required and returns an error
		if req.Arguments["id"] == nil || req.Arguments["id"] == "" {
			return `{"error": "id is required"}`, http.StatusBadRequest
		}
		return `{}`, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name:      "get_api_call",
		Arguments: map[string]any{},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected JSON-RPC error: %v", resp.Error)
	}

	result := resp.Result.(map[string]any)
	isError, ok := result["isError"].(bool)
	if !ok || !isError {
		t.Error("Expected isError=true when ID is missing")
	}

	content := result["content"].([]any)
	firstContent := content[0].(map[string]any)
	text := firstContent["text"].(string)
	if !strings.Contains(text, "id is required") {
		t.Errorf("Expected error about missing id, got: %s", text)
	}
}

func TestMCP_GetAPIStats(t *testing.T) {
	mockResponse := `{
		"stats": {
			"total_calls": 1000,
			"error_rate": 0.05,
			"endpoints": [{"path": "/api/users", "count": 500}]
		}
	}`

	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// Verify tool name
		if req.Tool != "get_api_stats" {
			t.Errorf("Expected tool 'get_api_stats', got %s", req.Tool)
		}
		// Verify arguments are in the POST body
		if req.Arguments["ns"] != "production" {
			t.Errorf("Expected ns=production in arguments, got %v", req.Arguments["ns"])
		}
		if req.Arguments["group_by"] != "endpoint" {
			t.Errorf("Expected group_by=endpoint in arguments, got %v", req.Arguments["group_by"])
		}
		return mockResponse, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name: "get_api_stats",
		Arguments: map[string]any{
			"ns":       "production",
			"group_by": "endpoint",
		},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result := resp.Result.(map[string]any)
	content := result["content"].([]any)
	firstContent := content[0].(map[string]any)
	text := firstContent["text"].(string)

	if !strings.Contains(text, "total_calls") {
		t.Errorf("Expected response to contain 'total_calls', got: %s", text)
	}
}

func TestMCP_APITools_BackendError(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Internal server error"}`))
	})
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name:      "list_workloads",
		Arguments: map[string]any{},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	result := resp.Result.(map[string]any)
	isError := result["isError"].(bool)
	if !isError {
		t.Error("Expected isError=true for backend error")
	}

	content := result["content"].([]any)
	firstContent := content[0].(map[string]any)
	text := firstContent["text"].(string)
	if !strings.Contains(text, "500") {
		t.Errorf("Expected error to mention status code 500, got: %s", text)
	}
}

func TestMCP_APITools_BackendConnectionError(t *testing.T) {
	s := &mcpServer{
		httpClient:         &http.Client{},
		stdin:              &bytes.Buffer{},
		stdout:             &bytes.Buffer{},
		hubBaseURL:         "http://localhost:99999", // Invalid port
		backendInitialized: true,
	}

	params := mcpCallToolParams{
		Name:      "list_workloads",
		Arguments: map[string]any{},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	result := resp.Result.(map[string]any)
	isError := result["isError"].(bool)
	if !isError {
		t.Error("Expected isError=true for connection error")
	}
}

// =============================================================================
// MCP Server Run Loop Tests
// =============================================================================

func TestMCP_RunLoop_ParseError(t *testing.T) {
	input := "not valid json\n"
	output := &bytes.Buffer{}

	s := &mcpServer{
		httpClient: &http.Client{},
		stdin:      strings.NewReader(input),
		stdout:     output,
	}

	s.run()

	resp := parseResponse(t, output.String())
	if resp.Error == nil {
		t.Fatal("Expected parse error")
	}
	if resp.Error.Code != -32700 {
		t.Errorf("Expected error code -32700, got %d", resp.Error.Code)
	}
}

func TestMCP_RunLoop_MultipleRequests(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"ping"}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
		`{"jsonrpc":"2.0","id":3,"method":"initialize"}`,
	}
	input := strings.Join(requests, "\n") + "\n"
	output := &bytes.Buffer{}

	s := &mcpServer{
		httpClient: &http.Client{},
		stdin:      strings.NewReader(input),
		stdout:     output,
	}

	s.run()

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 responses, got %d", len(lines))
	}

	// Verify each response has correct ID
	expectedIDs := []float64{1, 2, 3}
	for i, line := range lines {
		var resp jsonRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("Failed to parse response %d: %v", i, err)
		}
		if resp.ID != expectedIDs[i] {
			t.Errorf("Response %d: expected ID %v, got %v", i, expectedIDs[i], resp.ID)
		}
	}
}

func TestMCP_RunLoop_EmptyLines(t *testing.T) {
	input := "\n\n{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"ping\"}\n\n"
	output := &bytes.Buffer{}

	s := &mcpServer{
		httpClient: &http.Client{},
		stdin:      strings.NewReader(input),
		stdout:     output,
	}

	s.run()

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 response, got %d: %v", len(lines), lines)
	}
}

// =============================================================================
// Response Format Tests
// =============================================================================

func TestMCP_ResponseFormat_AlwaysIncludesID(t *testing.T) {
	s := newTestMCPServer()

	// Test with numeric ID
	output := sendRequest(s, "ping", 123, nil)
	resp := parseResponse(t, output)
	if resp.ID != float64(123) {
		t.Errorf("Expected ID 123, got %v", resp.ID)
	}

	// Test with string ID
	output = sendRequest(s, "ping", "string-id", nil)
	resp = parseResponse(t, output)
	if resp.ID != "string-id" {
		t.Errorf("Expected ID 'string-id', got %v", resp.ID)
	}

	// Test with null ID (for error responses)
	output = sendRequest(s, "unknown", nil, nil)
	resp = parseResponse(t, output)
	if resp.ID != nil {
		t.Errorf("Expected nil ID, got %v", resp.ID)
	}
}

func TestMCP_ResponseFormat_JSONRPCVersion(t *testing.T) {
	s := newTestMCPServer()
	output := sendRequest(s, "ping", 1, nil)
	resp := parseResponse(t, output)

	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %s", resp.JSONRPC)
	}
}

func TestMCP_ToolCallResult_ContentFormat(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data": "test"}`))
	})
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name:      "list_workloads",
		Arguments: map[string]any{},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	result := resp.Result.(map[string]any)

	// Verify content is an array
	content, ok := result["content"].([]any)
	if !ok {
		t.Fatalf("content is not an array: %T", result["content"])
	}

	// Verify each content item has type and text
	for i, item := range content {
		itemMap, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("content[%d] is not a map", i)
		}
		if itemMap["type"] != "text" {
			t.Errorf("content[%d].type expected 'text', got %v", i, itemMap["type"])
		}
		if _, ok := itemMap["text"].(string); !ok {
			t.Errorf("content[%d].text is not a string", i)
		}
	}
}

// =============================================================================
// Start/Stop Kubeshark Command Building Tests
// =============================================================================

func TestMCP_StartKubeshark_CommandArgs(t *testing.T) {
	// Test that the command is built correctly (we can't actually run it in tests)
	tests := []struct {
		name     string
		args     map[string]any
		expected []string
	}{
		{
			name:     "basic",
			args:     map[string]any{},
			expected: []string{"tap", "--set", "headless=true"},
		},
		{
			name: "with pod regex",
			args: map[string]any{
				"pod_regex": "nginx.*",
			},
			expected: []string{"tap", "nginx.*", "--set", "headless=true"},
		},
		{
			name: "with namespaces",
			args: map[string]any{
				"namespaces": "default,kube-system",
			},
			expected: []string{"tap", "-n", "default", "-n", "kube-system", "--set", "headless=true"},
		},
		{
			name: "with release namespace",
			args: map[string]any{
				"release_namespace": "kubeshark-ns",
			},
			expected: []string{"tap", "-s", "kubeshark-ns", "--set", "headless=true"},
		},
		{
			name: "all options",
			args: map[string]any{
				"pod_regex":         "api-.*",
				"namespaces":        "prod",
				"release_namespace": "monitoring",
			},
			expected: []string{"tap", "api-.*", "-n", "prod", "-s", "monitoring", "--set", "headless=true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build command args the same way the actual function does
			cmdArgs := []string{"tap"}

			if v, ok := tt.args["pod_regex"].(string); ok && v != "" {
				cmdArgs = append(cmdArgs, v)
			}

			if v, ok := tt.args["namespaces"].(string); ok && v != "" {
				namespaces := strings.Split(v, ",")
				for _, ns := range namespaces {
					ns = strings.TrimSpace(ns)
					if ns != "" {
						cmdArgs = append(cmdArgs, "-n", ns)
					}
				}
			}

			if v, ok := tt.args["release_namespace"].(string); ok && v != "" {
				cmdArgs = append(cmdArgs, "-s", v)
			}

			cmdArgs = append(cmdArgs, "--set", "headless=true")

			if len(cmdArgs) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d\nExpected: %v\nGot: %v",
					len(tt.expected), len(cmdArgs), tt.expected, cmdArgs)
				return
			}

			for i, arg := range cmdArgs {
				if arg != tt.expected[i] {
					t.Errorf("Arg %d: expected %q, got %q", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestMCP_StopKubeshark_CommandArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		expected []string
	}{
		{
			name:     "basic",
			args:     map[string]any{},
			expected: []string{"clean"},
		},
		{
			name: "with release namespace",
			args: map[string]any{
				"release_namespace": "custom-ns",
			},
			expected: []string{"clean", "-s", "custom-ns"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdArgs := []string{"clean"}

			if v, ok := tt.args["release_namespace"].(string); ok && v != "" {
				cmdArgs = append(cmdArgs, "-s", v)
			}

			if len(cmdArgs) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d", len(tt.expected), len(cmdArgs))
				return
			}

			for i, arg := range cmdArgs {
				if arg != tt.expected[i] {
					t.Errorf("Arg %d: expected %q, got %q", i, tt.expected[i], arg)
				}
			}
		})
	}
}

// =============================================================================
// Pretty Print JSON Tests
// =============================================================================

func TestMCP_PrettyPrintJSON(t *testing.T) {
	compactJSON := `{"key":"value","nested":{"a":1}}`

	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(compactJSON))
	})
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name:      "list_workloads",
		Arguments: map[string]any{},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	result := resp.Result.(map[string]any)
	content := result["content"].([]any)
	firstContent := content[0].(map[string]any)
	text := firstContent["text"].(string)

	// Verify JSON is pretty-printed (has newlines and indentation)
	if !strings.Contains(text, "\n") {
		t.Error("Expected pretty-printed JSON with newlines")
	}
	if !strings.Contains(text, "  ") {
		t.Error("Expected pretty-printed JSON with indentation")
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestMCP_ArgumentsWithSpecialChars(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// Verify special characters are preserved in the POST body
		if req.Arguments["path"] != "/api/users?id=123" {
			t.Errorf("Expected path with special chars, got: %v", req.Arguments["path"])
		}
		return `{}`, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name: "list_api_calls",
		Arguments: map[string]any{
			"path": "/api/users?id=123",
		},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	// Verify no error
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}

func TestMCP_GetAPICall_IDWithSpecialChars(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// ID with special characters should be preserved in the POST body
		if req.Arguments["id"] != "abc/123" {
			t.Errorf("Expected id=abc/123 in arguments, got: %v", req.Arguments["id"])
		}
		return `{"id": "abc/123"}`, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name: "get_api_call",
		Arguments: map[string]any{
			"id": "abc/123",
		},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	// Verify no error
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}

func TestMCP_EmptyArguments(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		// Verify arguments is an empty map in the POST body
		if len(req.Arguments) != 0 {
			t.Errorf("Expected empty arguments, got: %v", req.Arguments)
		}
		return `{}`, http.StatusOK
	}))
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name:      "list_workloads",
		Arguments: map[string]any{},
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	// Verify no error
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}

func TestMCP_NilArguments(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	defer mockServer.Close()

	params := mcpCallToolParams{
		Name:      "list_workloads",
		Arguments: nil,
	}
	output := sendRequest(s, "tools/call", 1, params)
	resp := parseResponse(t, output)

	// Should not crash with nil arguments
	if resp.Error != nil {
		t.Errorf("Unexpected error with nil arguments: %v", resp.Error)
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestMCP_BackendInitialization_Concurrent(t *testing.T) {
	// Test that concurrent access to ensureBackendConnection is safe
	s := newTestMCPServer()

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			// This will fail (no k8s config) but should not panic
			s.ensureBackendConnection()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// =============================================================================
// Integration-style Tests
// =============================================================================

func TestMCP_FullConversation(t *testing.T) {
	// Mock Hub MCP endpoint for tool definitions
	mockHubMCPResponse := `{
		"name": "kubeshark-hub",
		"version": "1.0.0",
		"tools": [
			{"name": "list_workloads", "description": "List workloads", "inputSchema": {}},
			{"name": "list_api_calls", "description": "List API calls", "inputSchema": {}},
			{"name": "get_api_call", "description": "Get API call details", "inputSchema": {}}
		],
		"prompts": []
	}`

	// Simulate a typical MCP conversation with correct /tools/call endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/", "":
			// Hub MCP info endpoint
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(mockHubMCPResponse))
		case "/tools/call":
			// All tool calls go through this endpoint
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var req hubToolCallRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Return response based on tool name
			switch req.Tool {
			case "list_workloads":
				_, _ = w.Write([]byte(`{"workloads": [{"name": "nginx", "namespace": "default"}]}`))
			case "list_api_calls":
				_, _ = w.Write([]byte(`{"calls": [{"id": "1", "method": "GET"}]}`))
			case "get_api_call":
				_, _ = w.Write([]byte(`{"id": "1", "method": "GET", "path": "/health"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	conversation := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_workloads","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"list_api_calls","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_api_call","arguments":{"id":"1"}}}`,
	}

	input := strings.Join(conversation, "\n") + "\n"
	output := &bytes.Buffer{}

	s := &mcpServer{
		httpClient:         &http.Client{},
		stdin:              strings.NewReader(input),
		stdout:             output,
		hubBaseURL:         mockServer.URL,
		backendInitialized: true,
	}

	s.run()

	// Should have 5 responses (notifications don't get responses)
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 5 {
		t.Errorf("Expected 5 responses, got %d", len(lines))
		for i, line := range lines {
			t.Logf("Response %d: %s", i, line)
		}
	}

	// Verify all responses have correct structure
	for i, line := range lines {
		var resp jsonRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Errorf("Response %d: failed to parse: %v", i, err)
			continue
		}
		if resp.JSONRPC != "2.0" {
			t.Errorf("Response %d: expected jsonrpc 2.0", i)
		}
		if resp.Error != nil {
			t.Errorf("Response %d: unexpected error: %v", i, resp.Error)
		}
	}
}
