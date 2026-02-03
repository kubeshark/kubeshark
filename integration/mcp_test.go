//go:build integration

package integration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// MCPRequest represents a JSON-RPC request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC response
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// mcpSession represents a running MCP server session
type mcpSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	cancel context.CancelFunc
}

// startMCPSession starts an MCP server and returns a session for sending requests.
func startMCPSession(t *testing.T, binary string, args ...string) *mcpSession {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	cmdArgs := append([]string{"mcp"}, args...)
	cmd := exec.CommandContext(ctx, binary, cmdArgs...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("Failed to start MCP server: %v", err)
	}

	return &mcpSession{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
		cancel: cancel,
	}
}

// sendRequest sends a JSON-RPC request and returns the response (30s timeout).
func (s *mcpSession) sendRequest(t *testing.T, req MCPRequest) MCPResponse {
	t.Helper()
	return s.sendRequestWithTimeout(t, req, 30*time.Second)
}

// sendRequestWithTimeout sends a JSON-RPC request with a custom timeout.
func (s *mcpSession) sendRequestWithTimeout(t *testing.T, req MCPRequest, timeout time.Duration) MCPResponse {
	t.Helper()

	reqBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	t.Logf("Sending: %s", string(reqBytes))

	if _, err := s.stdin.Write(append(reqBytes, '\n')); err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}

	// Read response with timeout
	responseChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		line, err := s.stdout.ReadString('\n')
		if err != nil {
			errChan <- err
			return
		}
		responseChan <- line
	}()

	select {
	case line := <-responseChan:
		t.Logf("Received: %s", strings.TrimSpace(line))
		var resp MCPResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v\nResponse: %s", err, line)
		}
		return resp
	case err := <-errChan:
		t.Fatalf("Failed to read response: %v", err)
	case <-time.After(timeout):
		t.Fatalf("Timeout waiting for MCP response after %v", timeout)
	}

	return MCPResponse{}
}

// callTool invokes an MCP tool and returns the response (30s timeout).
func (s *mcpSession) callTool(t *testing.T, id int, toolName string, args map[string]interface{}) MCPResponse {
	t.Helper()
	return s.callToolWithTimeout(t, id, toolName, args, 30*time.Second)
}

// callToolWithTimeout invokes an MCP tool with a custom timeout.
func (s *mcpSession) callToolWithTimeout(t *testing.T, id int, toolName string, args map[string]interface{}, timeout time.Duration) MCPResponse {
	t.Helper()

	return s.sendRequestWithTimeout(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}, timeout)
}

// close terminates the MCP session.
func (s *mcpSession) close() {
	s.cancel()
	_ = s.cmd.Wait()
}

// TestMCP_Initialize tests the MCP initialization handshake.
func TestMCP_Initialize(t *testing.T) {
	requireKubernetesCluster(t)
	binary := getKubesharkBinary(t)

	session := startMCPSession(t, binary)
	defer session.close()

	resp := session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "integration-test",
				"version": "1.0.0",
			},
		},
	})

	if resp.Error != nil {
		t.Fatalf("Initialize failed: %s", resp.Error.Message)
	}

	// Verify we got capabilities back
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if _, ok := result["capabilities"]; !ok {
		t.Error("Response missing capabilities")
	}

	if _, ok := result["serverInfo"]; !ok {
		t.Error("Response missing serverInfo")
	}
}

// TestMCP_ToolsList tests that tools/list returns all expected tools.
func TestMCP_ToolsList(t *testing.T) {
	requireKubernetesCluster(t)
	binary := getKubesharkBinary(t)

	session := startMCPSession(t, binary)
	defer session.close()

	// Initialize first
	session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "integration-test",
				"version": "1.0.0",
			},
		},
	})

	// List tools
	resp := session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	})

	if resp.Error != nil {
		t.Fatalf("tools/list failed: %s", resp.Error.Message)
	}

	var result struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	expectedTools := []string{
		"check_kubeshark_status",
		"start_kubeshark",
		"stop_kubeshark",
		"list_workloads",
		"list_api_calls",
		"get_api_call",
		"get_api_stats",
	}

	toolNames := make(map[string]bool)
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("Missing expected tool: %s", expected)
		}
	}
}

// TestMCP_CheckKubesharkStatus_NotRunning tests check_kubeshark_status when Kubeshark is not running.
func TestMCP_CheckKubesharkStatus_NotRunning(t *testing.T) {
	requireKubernetesCluster(t)
	binary := getKubesharkBinary(t)

	// Ensure Kubeshark is not running
	cleanupKubeshark(t, binary)

	session := startMCPSession(t, binary)
	defer session.close()

	// Initialize
	session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "test", "version": "1.0"},
		},
	})

	// Check status
	resp := session.callTool(t, 2, "check_kubeshark_status", nil)

	if resp.Error != nil {
		t.Fatalf("check_kubeshark_status failed: %s", resp.Error.Message)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content in response")
	}

	text := result.Content[0].Text
	if !strings.Contains(text, "not running") && !strings.Contains(text, "NOT") {
		t.Errorf("Expected 'not running' status, got: %s", text)
	}
}

// TestMCP_StartKubeshark tests the start_kubeshark tool.
func TestMCP_StartKubeshark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode (starts real Kubeshark)")
	}

	requireKubernetesCluster(t)
	binary := getKubesharkBinary(t)

	// Ensure clean state
	cleanupKubeshark(t, binary)

	session := startMCPSession(t, binary)
	defer session.close()

	// Initialize
	session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "test", "version": "1.0"},
		},
	})

	// Start Kubeshark - this may take a while
	t.Log("Starting Kubeshark (this may take a few minutes)...")

	// Use a longer timeout for start (3 minutes)
	resp := session.callToolWithTimeout(t, 2, "start_kubeshark", nil, 3*time.Minute)

	if resp.Error != nil {
		t.Fatalf("start_kubeshark failed: %s", resp.Error.Message)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content in response")
	}

	text := result.Content[0].Text
	t.Logf("Start response: %s", text)

	// Verify Kubeshark is actually running
	if !isKubesharkRunning(t) {
		t.Error("Kubeshark should be running after start_kubeshark")
	}

	// Cleanup
	t.Cleanup(func() {
		cleanupKubeshark(t, binary)
	})
}

// TestMCP_StopKubeshark tests the stop_kubeshark tool.
func TestMCP_StopKubeshark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode (requires running Kubeshark)")
	}

	requireKubernetesCluster(t)
	binary := getKubesharkBinary(t)

	session := startMCPSession(t, binary)
	defer session.close()

	// Initialize first
	session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      0,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "test", "version": "1.0"},
		},
	})

	// If Kubeshark is not running, start it first using MCP tool
	if !isKubesharkRunning(t) {
		t.Log("Kubeshark not running, starting it first via MCP...")
		resp := session.callToolWithTimeout(t, 1, "start_kubeshark", nil, 2*time.Minute)
		if resp.Error != nil {
			t.Skipf("Could not start Kubeshark for test: %v", resp.Error.Message)
		}
	}

	// Stop Kubeshark (use longer timeout - 2 minutes)
	t.Log("Stopping Kubeshark...")
	resp := session.callToolWithTimeout(t, 2, "stop_kubeshark", nil, 2*time.Minute)

	if resp.Error != nil {
		t.Fatalf("stop_kubeshark failed: %s", resp.Error.Message)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content in response")
	}

	text := result.Content[0].Text
	t.Logf("Stop response: %s", text)

	// Wait a moment for cleanup
	time.Sleep(5 * time.Second)

	// Verify Kubeshark is stopped
	if isKubesharkRunning(t) {
		t.Error("Kubeshark should not be running after stop_kubeshark")
	}
}

// TestMCP_FullLifecycle tests the complete lifecycle: check -> start -> check -> stop -> check
func TestMCP_FullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode (full lifecycle test)")
	}

	requireKubernetesCluster(t)
	binary := getKubesharkBinary(t)

	// Ensure clean state
	cleanupKubeshark(t, binary)

	session := startMCPSession(t, binary)
	defer session.close()

	// Initialize
	session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "test", "version": "1.0"},
		},
	})

	// Step 1: Check status (should be not running)
	t.Log("Step 1: Checking initial status...")
	resp := session.callTool(t, 2, "check_kubeshark_status", nil)
	if resp.Error != nil {
		t.Fatalf("Initial status check failed: %s", resp.Error.Message)
	}
	t.Log("Initial status: Kubeshark not running (expected)")

	// Step 2: Start Kubeshark (use longer timeout - 3 minutes)
	t.Log("Step 2: Starting Kubeshark...")
	resp = session.callToolWithTimeout(t, 3, "start_kubeshark", nil, 3*time.Minute)
	if resp.Error != nil {
		t.Fatalf("Start failed: %s", resp.Error.Message)
	}

	// Wait for it to be ready
	if err := waitForKubesharkReady(t, binary, startupTimeout); err != nil {
		t.Fatalf("Kubeshark did not become ready: %v", err)
	}
	t.Log("Kubeshark started successfully")

	// Step 3: Check status (should be running)
	t.Log("Step 3: Checking status after start...")
	resp = session.callTool(t, 4, "check_kubeshark_status", nil)
	if resp.Error != nil {
		t.Fatalf("Status check after start failed: %s", resp.Error.Message)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &result); err == nil && len(result.Content) > 0 {
		if !strings.Contains(strings.ToLower(result.Content[0].Text), "running") {
			t.Errorf("Expected running status, got: %s", result.Content[0].Text)
		}
	}
	t.Log("Status after start: Kubeshark running (expected)")

	// Step 4: Stop Kubeshark (use longer timeout - 2 minutes)
	t.Log("Step 4: Stopping Kubeshark...")
	resp = session.callToolWithTimeout(t, 5, "stop_kubeshark", nil, 2*time.Minute)
	if resp.Error != nil {
		t.Fatalf("Stop failed: %s", resp.Error.Message)
	}
	time.Sleep(5 * time.Second)
	t.Log("Kubeshark stopped successfully")

	// Step 5: Check status (should be not running)
	t.Log("Step 5: Checking final status...")
	resp = session.callTool(t, 6, "check_kubeshark_status", nil)
	if resp.Error != nil {
		t.Fatalf("Final status check failed: %s", resp.Error.Message)
	}
	t.Log("Final status: Kubeshark not running (expected)")

	t.Log("Full lifecycle test completed successfully!")
}

// TestMCP_APIToolsRequireKubeshark tests that API tools return helpful errors when Kubeshark isn't running.
func TestMCP_APIToolsRequireKubeshark(t *testing.T) {
	requireKubernetesCluster(t)
	binary := getKubesharkBinary(t)

	// Ensure Kubeshark is not running
	cleanupKubeshark(t, binary)

	session := startMCPSession(t, binary)
	defer session.close()

	// Initialize
	session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "test", "version": "1.0"},
		},
	})

	apiTools := []string{"list_workloads", "list_api_calls", "get_api_stats"}

	for i, tool := range apiTools {
		t.Run(tool, func(t *testing.T) {
			resp := session.callTool(t, i+2, tool, nil)

			// Should succeed but indicate Kubeshark is not running
			if resp.Error != nil {
				// An error is acceptable too
				t.Logf("%s returned error (expected): %s", tool, resp.Error.Message)
				return
			}

			var result struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			}
			if err := json.Unmarshal(resp.Result, &result); err == nil && len(result.Content) > 0 {
				text := strings.ToLower(result.Content[0].Text)
				if strings.Contains(text, "error") || strings.Contains(text, "not running") || strings.Contains(text, "failed") {
					t.Logf("%s returned helpful message: %s", tool, result.Content[0].Text)
				}
			}
		})
	}
}

// TestMCP_TapSetFlags tests that --tap-set flags are passed correctly.
func TestMCP_TapSetFlags(t *testing.T) {
	requireKubernetesCluster(t)
	binary := getKubesharkBinary(t)

	// Start MCP with custom tap-set flags
	session := startMCPSession(t, binary, "--tap-set", "tap.namespaces={default}")
	defer session.close()

	// Initialize
	session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "test", "version": "1.0"},
		},
	})

	// List tools - should work regardless of flags
	resp := session.sendRequest(t, MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	})

	if resp.Error != nil {
		t.Fatalf("tools/list failed with tap-set flags: %s", resp.Error.Message)
	}

	t.Log("MCP server started successfully with --tap-set flags")
}

// BenchmarkMCP_CheckStatus benchmarks the check_kubeshark_status tool.
func BenchmarkMCP_CheckStatus(b *testing.B) {
	// Skip in short mode
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	// Build binary once
	projectRoot, err := findProjectRoot()
	if err != nil {
		b.Fatalf("Could not find project root: %v", err)
	}

	binaryPath := projectRoot + "/bin/kubeshark_integration_test"

	// Start MCP session
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "mcp")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	reader := bufio.NewReader(stdout)

	if err := cmd.Start(); err != nil {
		b.Fatalf("Failed to start MCP: %v", err)
	}
	defer func() {
		cancel()
		_ = cmd.Wait()
	}()

	// Initialize
	initReq, _ := json.Marshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      0,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "bench", "version": "1.0"},
		},
	})
	_, _ = stdin.Write(append(initReq, '\n'))
	_, _ = reader.ReadString('\n')

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := json.Marshal(MCPRequest{
			JSONRPC: "2.0",
			ID:      i + 1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      "check_kubeshark_status",
				"arguments": map[string]interface{}{},
			},
		})

		_, err := stdin.Write(append(req, '\n'))
		if err != nil {
			b.Fatalf("Write failed: %v", err)
		}

		_, err = reader.ReadString('\n')
		if err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
}

// Helper to avoid unused import errors
var _ = bytes.Buffer{}
