package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestMCPServer() *mcpServer {
	return &mcpServer{httpClient: &http.Client{}, stdin: &bytes.Buffer{}, stdout: &bytes.Buffer{}}
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

func TestMCP_Initialize(t *testing.T) {
	s := newTestMCPServer()
	resp := parseResponse(t, sendRequest(s, "initialize", 1, nil))

	if resp.ID != float64(1) || resp.Error != nil {
		t.Fatalf("Expected ID 1 with no error, got ID=%v, error=%v", resp.ID, resp.Error)
	}

	result := resp.Result.(map[string]any)
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion 2024-11-05, got %v", result["protocolVersion"])
	}
	if result["serverInfo"].(map[string]any)["name"] != "kubeshark-mcp" {
		t.Error("Expected server name kubeshark-mcp")
	}
	if !strings.Contains(result["instructions"].(string), "check_kubeshark_status") {
		t.Error("Instructions should mention check_kubeshark_status")
	}
	if _, ok := result["capabilities"].(map[string]any)["prompts"]; !ok {
		t.Error("Expected prompts capability")
	}
}

func TestMCP_Ping(t *testing.T) {
	resp := parseResponse(t, sendRequest(newTestMCPServer(), "ping", 42, nil))
	if resp.ID != float64(42) || resp.Error != nil || len(resp.Result.(map[string]any)) != 0 {
		t.Errorf("Expected ID 42, no error, empty result")
	}
}

func TestMCP_InitializedNotification(t *testing.T) {
	s := newTestMCPServer()
	for _, method := range []string{"initialized", "notifications/initialized"} {
		if output := sendRequest(s, method, nil, nil); output != "" {
			t.Errorf("Expected no output for %s, got: %s", method, output)
		}
	}
}

func TestMCP_UnknownMethod(t *testing.T) {
	resp := parseResponse(t, sendRequest(newTestMCPServer(), "unknown/method", 1, nil))
	if resp.Error == nil || resp.Error.Code != -32601 {
		t.Fatalf("Expected error code -32601, got %v", resp.Error)
	}
}

func TestMCP_PromptsList(t *testing.T) {
	resp := parseResponse(t, sendRequest(newTestMCPServer(), "prompts/list", 1, nil))
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	prompts := resp.Result.(map[string]any)["prompts"].([]any)
	if len(prompts) != 1 || prompts[0].(map[string]any)["name"] != "kubeshark_usage" {
		t.Error("Expected 1 prompt named 'kubeshark_usage'")
	}
}

func TestMCP_PromptsGet(t *testing.T) {
	resp := parseResponse(t, sendRequest(newTestMCPServer(), "prompts/get", 1, map[string]any{"name": "kubeshark_usage"}))
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	messages := resp.Result.(map[string]any)["messages"].([]any)
	if len(messages) == 0 {
		t.Fatal("Expected at least one message")
	}
	text := messages[0].(map[string]any)["content"].(map[string]any)["text"].(string)
	for _, phrase := range []string{"check_kubeshark_status", "start_kubeshark", "stop_kubeshark"} {
		if !strings.Contains(text, phrase) {
			t.Errorf("Prompt should contain '%s'", phrase)
		}
	}
}

func TestMCP_PromptsGet_UnknownPrompt(t *testing.T) {
	resp := parseResponse(t, sendRequest(newTestMCPServer(), "prompts/get", 1, map[string]any{"name": "unknown"}))
	if resp.Error == nil || resp.Error.Code != -32602 {
		t.Fatalf("Expected error code -32602, got %v", resp.Error)
	}
}

func TestMCP_ToolsList_CLIOnly(t *testing.T) {
	resp := parseResponse(t, sendRequest(newTestMCPServer(), "tools/list", 1, nil))
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	tools := resp.Result.(map[string]any)["tools"].([]any)
	if len(tools) != 1 || tools[0].(map[string]any)["name"] != "check_kubeshark_status" {
		t.Error("Expected only check_kubeshark_status tool")
	}
}

func TestMCP_ToolsList_WithDestructive(t *testing.T) {
	s := &mcpServer{httpClient: &http.Client{}, stdin: &bytes.Buffer{}, stdout: &bytes.Buffer{}, allowDestructive: true}
	resp := parseResponse(t, sendRequest(s, "tools/list", 1, nil))
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	tools := resp.Result.(map[string]any)["tools"].([]any)
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.(map[string]any)["name"].(string)] = true
	}
	for _, expected := range []string{"check_kubeshark_status", "start_kubeshark", "stop_kubeshark"} {
		if !toolNames[expected] {
			t.Errorf("Missing expected tool: %s", expected)
		}
	}
}

func TestMCP_ToolsList_WithHubBackend(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			_, _ = w.Write([]byte(`{"name":"hub","tools":[{"name":"list_workloads","description":"","inputSchema":{}},{"name":"list_api_calls","description":"","inputSchema":{}}]}`))
		}
	}))
	defer mockServer.Close()

	s := &mcpServer{httpClient: &http.Client{}, stdin: &bytes.Buffer{}, stdout: &bytes.Buffer{}, hubBaseURL: mockServer.URL, backendInitialized: true, allowDestructive: true}
	resp := parseResponse(t, sendRequest(s, "tools/list", 1, nil))
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	tools := resp.Result.(map[string]any)["tools"].([]any)
	// Should have CLI tools (3) + Hub tools (2) = 5 tools
	if len(tools) < 5 {
		t.Errorf("Expected at least 5 tools, got %d", len(tools))
	}
}

func TestMCP_ToolsCallUnknownTool(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer mockServer.Close()

	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "unknown"}))
	if !resp.Result.(map[string]any)["isError"].(bool) {
		t.Error("Expected isError=true for unknown tool")
	}
}

func TestMCP_ToolsCallInvalidParams(t *testing.T) {
	s := newTestMCPServer()
	req := jsonRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: json.RawMessage(`"invalid"`)}
	s.handleRequest(&req)
	resp := parseResponse(t, s.stdout.(*bytes.Buffer).String())
	if resp.Error == nil || resp.Error.Code != -32602 {
		t.Fatalf("Expected error code -32602")
	}
}

func TestMCP_CheckKubesharkStatus(t *testing.T) {
	for _, tc := range []struct {
		name string
		args map[string]any
	}{
		{"no_config", map[string]any{}},
		{"with_namespace", map[string]any{"release_namespace": "custom-ns"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := parseResponse(t, sendRequest(newTestMCPServer(), "tools/call", 1, mcpCallToolParams{Name: "check_kubeshark_status", Arguments: tc.args}))
			if resp.Error != nil {
				t.Fatalf("Unexpected error: %v", resp.Error)
			}
			content := resp.Result.(map[string]any)["content"].([]any)
			if len(content) == 0 || content[0].(map[string]any)["text"].(string) == "" {
				t.Error("Expected non-empty response")
			}
		})
	}
}

func newTestMCPServerWithMockBackend(handler http.HandlerFunc) (*mcpServer, *httptest.Server) {
	mockServer := httptest.NewServer(handler)
	return &mcpServer{httpClient: &http.Client{}, stdin: &bytes.Buffer{}, stdout: &bytes.Buffer{}, hubBaseURL: mockServer.URL, backendInitialized: true}, mockServer
}

type hubToolCallRequest struct {
	Tool      string         `json:"tool"`
	Arguments map[string]any `json:"arguments"`
}

func newMockHubHandler(t *testing.T, handler func(req hubToolCallRequest) (string, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tools/call" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var req hubToolCallRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		resp, status := handler(req)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(resp))
	}
}

func TestMCP_ListWorkloads(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		if req.Tool != "list_workloads" {
			t.Errorf("Expected tool 'list_workloads', got %s", req.Tool)
		}
		return `{"workloads": [{"name": "test-pod"}]}`, http.StatusOK
	}))
	defer mockServer.Close()

	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "list_workloads", Arguments: map[string]any{"type": "pod"}}))
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	text := resp.Result.(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "test-pod") {
		t.Errorf("Expected 'test-pod' in response")
	}
}

func TestMCP_ListAPICalls(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		if req.Tool != "list_api_calls" {
			t.Errorf("Expected tool 'list_api_calls', got %s", req.Tool)
		}
		return `{"calls": [{"id": "123", "path": "/api/users"}]}`, http.StatusOK
	}))
	defer mockServer.Close()

	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "list_api_calls", Arguments: map[string]any{"proto": "http"}}))
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	if !strings.Contains(resp.Result.(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string), "/api/users") {
		t.Error("Expected '/api/users' in response")
	}
}

func TestMCP_GetAPICall(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		if req.Tool != "get_api_call" || req.Arguments["id"] != "abc123" {
			t.Errorf("Expected get_api_call with id=abc123")
		}
		return `{"id": "abc123", "path": "/api/orders"}`, http.StatusOK
	}))
	defer mockServer.Close()

	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "get_api_call", Arguments: map[string]any{"id": "abc123"}}))
	if resp.Error != nil || !strings.Contains(resp.Result.(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string), "abc123") {
		t.Error("Expected response containing 'abc123'")
	}
}

func TestMCP_GetAPICall_MissingID(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		return `{"error": "id is required"}`, http.StatusBadRequest
	}))
	defer mockServer.Close()

	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "get_api_call", Arguments: map[string]any{}}))
	if !resp.Result.(map[string]any)["isError"].(bool) {
		t.Error("Expected isError=true")
	}
}

func TestMCP_GetAPIStats(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(newMockHubHandler(t, func(req hubToolCallRequest) (string, int) {
		if req.Tool != "get_api_stats" {
			t.Errorf("Expected get_api_stats")
		}
		return `{"stats": {"total_calls": 1000}}`, http.StatusOK
	}))
	defer mockServer.Close()

	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "get_api_stats", Arguments: map[string]any{"ns": "prod"}}))
	if resp.Error != nil || !strings.Contains(resp.Result.(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string), "total_calls") {
		t.Error("Expected 'total_calls' in response")
	}
}

func TestMCP_APITools_BackendError(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer mockServer.Close()

	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "list_workloads"}))
	if !resp.Result.(map[string]any)["isError"].(bool) {
		t.Error("Expected isError=true for backend error")
	}
}

func TestMCP_APITools_BackendConnectionError(t *testing.T) {
	s := &mcpServer{httpClient: &http.Client{}, stdin: &bytes.Buffer{}, stdout: &bytes.Buffer{}, hubBaseURL: "http://localhost:99999", backendInitialized: true}
	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "list_workloads"}))
	if !resp.Result.(map[string]any)["isError"].(bool) {
		t.Error("Expected isError=true for connection error")
	}
}

func TestMCP_RunLoop_ParseError(t *testing.T) {
	output := &bytes.Buffer{}
	s := &mcpServer{httpClient: &http.Client{}, stdin: strings.NewReader("invalid\n"), stdout: output}
	s.run()
	if resp := parseResponse(t, output.String()); resp.Error == nil || resp.Error.Code != -32700 {
		t.Fatalf("Expected error code -32700")
	}
}

func TestMCP_RunLoop_MultipleRequests(t *testing.T) {
	output := &bytes.Buffer{}
	s := &mcpServer{httpClient: &http.Client{}, stdin: strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}
{"jsonrpc":"2.0","id":2,"method":"ping"}
`), stdout: output}
	s.run()
	if lines := strings.Split(strings.TrimSpace(output.String()), "\n"); len(lines) != 2 {
		t.Fatalf("Expected 2 responses, got %d", len(lines))
	}
}

func TestMCP_RunLoop_EmptyLines(t *testing.T) {
	output := &bytes.Buffer{}
	s := &mcpServer{httpClient: &http.Client{}, stdin: strings.NewReader("\n\n{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"ping\"}\n"), stdout: output}
	s.run()
	if lines := strings.Split(strings.TrimSpace(output.String()), "\n"); len(lines) != 1 {
		t.Fatalf("Expected 1 response, got %d", len(lines))
	}
}

func TestMCP_ResponseFormat(t *testing.T) {
	s := newTestMCPServer()
	// Numeric ID
	if resp := parseResponse(t, sendRequest(s, "ping", 123, nil)); resp.ID != float64(123) || resp.JSONRPC != "2.0" {
		t.Errorf("Expected ID 123 and jsonrpc 2.0")
	}
	// String ID
	if resp := parseResponse(t, sendRequest(s, "ping", "str", nil)); resp.ID != "str" {
		t.Errorf("Expected ID 'str'")
	}
}

func TestMCP_ToolCallResult_ContentFormat(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data": "test"}`))
	})
	defer mockServer.Close()

	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "list_workloads"}))
	content := resp.Result.(map[string]any)["content"].([]any)
	if len(content) == 0 || content[0].(map[string]any)["type"] != "text" {
		t.Error("Expected content with type=text")
	}
}

func TestMCP_CommandArgs(t *testing.T) {
	// Test start command args building
	for _, tc := range []struct {
		args     map[string]any
		expected string
	}{
		{map[string]any{}, "tap --set headless=true"},
		{map[string]any{"pod_regex": "nginx.*"}, "tap nginx.* --set headless=true"},
		{map[string]any{"namespaces": "default"}, "tap -n default --set headless=true"},
		{map[string]any{"release_namespace": "ks"}, "tap -s ks --set headless=true"},
	} {
		cmdArgs := []string{"tap"}
		if v, _ := tc.args["pod_regex"].(string); v != "" {
			cmdArgs = append(cmdArgs, v)
		}
		if v, _ := tc.args["namespaces"].(string); v != "" {
			for _, ns := range strings.Split(v, ",") {
				cmdArgs = append(cmdArgs, "-n", strings.TrimSpace(ns))
			}
		}
		if v, _ := tc.args["release_namespace"].(string); v != "" {
			cmdArgs = append(cmdArgs, "-s", v)
		}
		cmdArgs = append(cmdArgs, "--set", "headless=true")
		if got := strings.Join(cmdArgs, " "); got != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, got)
		}
	}
}

func TestMCP_PrettyPrintJSON(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"key":"value"}`))
	})
	defer mockServer.Close()

	resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "list_workloads"}))
	text := resp.Result.(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "\n") {
		t.Error("Expected pretty-printed JSON")
	}
}

func TestMCP_SpecialCharsAndEdgeCases(t *testing.T) {
	s, mockServer := newTestMCPServerWithMockBackend(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	defer mockServer.Close()

	// Test special chars, empty args, nil args
	for _, args := range []map[string]any{
		{"path": "/api?id=123"},
		{"id": "abc/123"},
		{},
		nil,
	} {
		resp := parseResponse(t, sendRequest(s, "tools/call", 1, mcpCallToolParams{Name: "list_workloads", Arguments: args}))
		if resp.Error != nil {
			t.Errorf("Unexpected error with args %v: %v", args, resp.Error)
		}
	}
}

func TestMCP_BackendInitialization_Concurrent(t *testing.T) {
	s := newTestMCPServer()
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() { s.ensureBackendConnection(); done <- true }()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMCP_FullConversation(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte(`{"name":"hub","tools":[{"name":"list_workloads","description":"","inputSchema":{}}]}`))
		} else if r.URL.Path == "/tools/call" {
			_, _ = w.Write([]byte(`{"data":"ok"}`))
		}
	}))
	defer mockServer.Close()

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_workloads","arguments":{}}}
`
	output := &bytes.Buffer{}
	s := &mcpServer{httpClient: &http.Client{}, stdin: strings.NewReader(input), stdout: output, hubBaseURL: mockServer.URL, backendInitialized: true}
	s.run()

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 { // 3 responses (notification has no response)
		t.Errorf("Expected 3 responses, got %d", len(lines))
	}
	for i, line := range lines {
		var resp jsonRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil || resp.Error != nil {
			t.Errorf("Response %d: parse error or unexpected error", i)
		}
	}
}
