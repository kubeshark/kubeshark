# Add Integration Test Framework

## Overview

Create an integration test framework that runs against real infrastructure (not mocks). Tests should be separate from unit tests and only run when explicitly requested.

## Requirements

### Directory Structure

```
integration/
├── README.md           # Documentation for running tests
├── common_test.go      # Shared test helpers
└── <feature>_test.go   # Tests for specific features/commands
```

### Makefile Targets

Add to existing Makefile:

```makefile
test-integration: ## Run integration tests
	go test -tags=integration -timeout $${INTEGRATION_TIMEOUT:-5m} ./integration/...

test-integration-verbose: ## Run integration tests with verbose output
	go test -tags=integration -timeout $${INTEGRATION_TIMEOUT:-5m} -v ./integration/...

test-integration-short: ## Run quick integration tests (skip long-running)
	go test -tags=integration -timeout $${INTEGRATION_TIMEOUT:-2m} -short -v ./integration/...
```

### common_test.go Requirements

Create shared helpers:

1. **`requirePrerequisites(t)`** - Skip test if infrastructure not available
2. **`getBinary(t)`** - Build or locate the binary (cache across tests)
3. **`runCommand(t, binary, args...)`** - Execute with timeout, return output
4. **`runCommandWithTimeout(t, binary, timeout, args...)`** - Custom timeout variant
5. **`cleanup(t, binary)`** - Clean up resources after tests

Key patterns:
- All files must have `//go:build integration` tag
- Use `sync.Once` for binary building (build once per test run)
- Support `TEST_BINARY` env var for pre-built binary
- Support `INTEGRATION_TIMEOUT` env var for custom timeout
- Support `INTEGRATION_SKIP_CLEANUP` env var for debugging
- Use `t.Skip()` when prerequisites aren't met
- Use `testing.Short()` to skip long-running tests

### Test File Requirements

Each test file should:
- Have `//go:build integration` tag
- Call `requirePrerequisites(t)` first
- Call `cleanup(t, binary)` to ensure clean state
- Use `t.Logf()` for progress visibility
- Mark long-running tests to skip in short mode

Example structure:
```go
//go:build integration

package integration

func TestFeature_Scenario(t *testing.T) {
    requirePrerequisites(t)
    binary := getBinary(t)
    cleanup(t, binary)

    // Test logic
}

func TestFeature_LongRunning(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }
    // ...
}
```

### README.md Requirements

Document:
- Prerequisites (what infrastructure is needed)
- How to run tests (`make test-integration`, etc.)
- Environment variables table
- How to write new tests

## Acceptance Criteria

- [ ] `make test-integration` runs all integration tests
- [ ] `make test-integration-short` skips long-running tests
- [ ] Tests skip gracefully when prerequisites unavailable
- [ ] Tests clean up resources after completion
- [ ] README documents how to run and write tests
- [ ] All tests pass on a properly configured environment

## Reference

See working example: https://github.com/kubeshark/kubeshark/tree/master/integration
