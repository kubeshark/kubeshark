# Integration Test Framework Template

This document describes a reusable integration test pattern for Go CLI projects. Reference this file when setting up integration tests in a new repository.

## Usage

Tell Claude: "Read /path/to/TEMPLATE.md and implement the same integration test framework for this repo"

---

## Directory Structure

```
your-repo/
├── integration/
│   ├── README.md           # Documentation for running tests
│   ├── common_test.go      # Shared test helpers (adapt for your project)
│   └── <feature>_test.go   # Tests for specific features/commands
└── Makefile                # Add integration test targets
```

---

## File Templates

### 1. Makefile Targets

Add these to your existing Makefile:

```makefile
test-integration: ## Run integration tests
	go test -tags=integration -timeout $${INTEGRATION_TIMEOUT:-5m} ./integration/...

test-integration-verbose: ## Run integration tests with verbose output
	go test -tags=integration -timeout $${INTEGRATION_TIMEOUT:-5m} -v ./integration/...

test-integration-short: ## Run quick integration tests (skip long-running)
	go test -tags=integration -timeout $${INTEGRATION_TIMEOUT:-2m} -short -v ./integration/...
```

### 2. common_test.go Template

```go
//go:build integration

// Package integration contains integration tests that run against real infrastructure.
// Run with: go test -tags=integration ./integration/...
package integration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	// ADAPT: Change to your binary name
	binaryName = "your-binary"

	// Timeouts
	defaultTimeout = 2 * time.Minute
	startupTimeout = 3 * time.Minute
)

var (
	binaryPath string
	buildOnce  sync.Once
	buildErr   error
)

// ADAPT: Check for your prerequisites (database, cluster, service, etc.)
func requirePrerequisites(t *testing.T) {
	t.Helper()

	// Example: Check for Kubernetes cluster
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", "cluster-info")
	if err := cmd.Run(); err != nil {
		t.Skip("Skipping: prerequisites not available")
	}
}

// getBinary returns path to the binary, building if necessary.
func getBinary(t *testing.T) string {
	t.Helper()

	// Check environment variable first
	if envBinary := os.Getenv("TEST_BINARY"); envBinary != "" {
		if _, err := os.Stat(envBinary); err == nil {
			return envBinary
		}
		t.Fatalf("TEST_BINARY set but not found: %s", envBinary)
	}

	// Build once per test run
	buildOnce.Do(func() {
		binaryPath, buildErr = buildBinary(t)
	})

	if buildErr != nil {
		t.Fatalf("Failed to build binary: %v", buildErr)
	}
	return binaryPath
}

// ADAPT: Modify build command for your project
func buildBinary(t *testing.T) (string, error) {
	t.Helper()

	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}

	outputPath := filepath.Join(projectRoot, "bin", binaryName+"_integration_test")
	t.Logf("Building %s at %s", binaryName, outputPath)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// ADAPT: Change build command as needed
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outputPath, ".")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %w\nOutput: %s", err, output)
	}
	return outputPath, nil
}

func findProjectRoot() (string, error) {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// runCommand executes a command with timeout.
func runCommand(t *testing.T, binary string, args ...string) (string, error) {
	t.Helper()
	return runCommandWithTimeout(t, binary, defaultTimeout, args...)
}

func runCommandWithTimeout(t *testing.T, binary string, timeout time.Duration, args ...string) (string, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t.Logf("Running: %s %s", binary, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n[stderr]\n" + stderr.String()
	}

	if ctx.Err() == context.DeadlineExceeded {
		return output, fmt.Errorf("timeout after %v", timeout)
	}
	return output, err
}

// ADAPT: Implement cleanup for your project's resources
func cleanup(t *testing.T, binary string) {
	t.Helper()
	if os.Getenv("INTEGRATION_SKIP_CLEANUP") == "true" {
		t.Log("Skipping cleanup")
		return
	}
	t.Log("Cleaning up...")
	// Example: runCommand(t, binary, "cleanup", "--force")
}
```

### 3. README.md Template

```markdown
# Integration Tests

Integration tests that run against real infrastructure.

## Prerequisites

- [List your prerequisites: K8s cluster, database, etc.]
- Go 1.21+

## Running Tests

\`\`\`bash
# All tests
make test-integration

# Verbose output
make test-integration-verbose

# Quick tests only
make test-integration-short
\`\`\`

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TEST_BINARY` | Auto-built | Path to pre-built binary |
| `INTEGRATION_TIMEOUT` | `5m` | Test timeout |
| `INTEGRATION_SKIP_CLEANUP` | `false` | Skip cleanup (debugging) |

## Writing Tests

1. Create `<feature>_test.go`
2. Add build tag: `//go:build integration`
3. Use helpers from `common_test.go`

\`\`\`go
//go:build integration

package integration

func TestMyFeature(t *testing.T) {
    requirePrerequisites(t)
    binary := getBinary(t)
    cleanup(t, binary)

    // Test logic here
    output, err := runCommand(t, binary, "my-command", "--flag")
    if err != nil {
        t.Fatalf("Failed: %v\nOutput: %s", err, output)
    }
}
\`\`\`
```

### 4. Example Test File

```go
//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestFeature_BasicUsage(t *testing.T) {
	requirePrerequisites(t)
	binary := getBinary(t)

	output, err := runCommand(t, binary, "help")
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage info, got: %s", output)
	}
}

func TestFeature_LongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	requirePrerequisites(t)
	binary := getBinary(t)
	cleanup(t, binary)

	// Long-running test logic
}
```

---

## Adaptation Checklist

When implementing in a new repo:

1. [ ] Change `binaryName` constant
2. [ ] Update `buildBinary()` with correct build command
3. [ ] Implement `requirePrerequisites()` for your infrastructure
4. [ ] Implement `cleanup()` for your resources
5. [ ] Add Makefile targets
6. [ ] Create feature-specific test files
7. [ ] Update README with your prerequisites

---

## Key Patterns

1. **Build tag**: Always use `//go:build integration` to separate from unit tests
2. **Skip gracefully**: Use `t.Skip()` when prerequisites aren't met
3. **Cleanup always**: Clean up resources even on failure
4. **Timeouts**: Use context with timeouts for all external calls
5. **Short mode**: Use `testing.Short()` to skip long-running tests
6. **Logging**: Use `t.Logf()` for progress visibility
