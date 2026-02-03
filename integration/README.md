# Integration Tests

This directory contains integration tests that run against a real Kubernetes cluster.

## Prerequisites

1. **Kubernetes cluster** - A running cluster accessible via `kubectl`
2. **kubectl** - Configured with appropriate context
3. **Go 1.21+** - For running tests

## Running Tests

```bash
# Run all integration tests
make test-integration

# Run specific command tests
make test-integration-mcp

# Run with verbose output
make test-integration-verbose

# Run with custom timeout (default: 5m)
INTEGRATION_TIMEOUT=10m make test-integration
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `KUBESHARK_BINARY` | Auto-built | Path to pre-built kubeshark binary |
| `INTEGRATION_TIMEOUT` | `5m` | Test timeout duration |
| `KUBECONFIG` | `~/.kube/config` | Kubernetes config file |
| `INTEGRATION_SKIP_CLEANUP` | `false` | Skip cleanup after tests (for debugging) |

## Test Structure

```
integration/
├── README.md           # This file
├── common_test.go      # Shared test helpers
├── mcp_test.go         # MCP command integration tests
├── tap_test.go         # Tap command tests (future)
└── ...                 # Additional command tests
```

## Writing New Tests

1. Create a new file `<command>_test.go`
2. Use the build tag `//go:build integration`
3. Use helpers from `common_test.go`

Example:
```go
//go:build integration

package integration

import (
    "testing"
)

func TestMyCommand_Feature(t *testing.T) {
    // Ensure cluster is available
    requireKubernetesCluster(t)

    // Get binary path (auto-builds if needed)
    binary := getKubesharkBinary(t)

    // Ensure clean state
    cleanupKubeshark(t, binary)

    // Your test logic here
    output, err := runKubeshark(t, binary, "my-command", "--flag")
    if err != nil {
        t.Fatalf("command failed: %v\nOutput: %s", err, output)
    }
}
```

## CI/CD Integration

Tests output standard Go test format, compatible with most CI systems:

```bash
# JSON output for CI parsing
go test -tags=integration -json ./integration/...

# With coverage
go test -tags=integration -coverprofile=integration-coverage.out ./integration/...
```

---

## Reusable Template for Other Repos

This integration test framework follows a pattern that can be applied to any Go project. To use this pattern in another repository:

### 1. Copy the structure

```bash
mkdir -p integration
# Copy common_test.go as a starting template
# Adapt the binary name and commands to your project
```

### 2. Adapt common_test.go

Replace these project-specific items:
- Binary name (`kubeshark` -> your binary)
- Build command (`go build` path)
- Cleanup logic (what resources to clean up)
- Health check logic (how to verify your service is running)

### 3. Add Makefile targets

```makefile
test-integration: ## Run integration tests
	go test -tags=integration -timeout $${INTEGRATION_TIMEOUT:-5m} ./integration/...

test-integration-verbose: ## Run integration tests with verbose output
	go test -tags=integration -timeout $${INTEGRATION_TIMEOUT:-5m} -v ./integration/...
```

### 4. Key patterns to follow

1. **Build tag**: Always use `//go:build integration`
2. **Skip gracefully**: Use `t.Skip()` when prerequisites aren't met
3. **Cleanup**: Always clean up resources, even on failure
4. **Timeouts**: Use context with timeouts for all external calls
5. **Parallel safety**: Don't assume tests run in order
