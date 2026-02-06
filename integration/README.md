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

1. Create `<command>_test.go` with build tag `//go:build integration`
2. Use helpers from `common_test.go`: `requireKubernetesCluster(t)`, `getKubesharkBinary(t)`, `cleanupKubeshark(t, binary)`

## CI/CD Integration

```bash
# JSON output for CI parsing
go test -tags=integration -json ./integration/...
```
