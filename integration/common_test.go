//go:build integration

// Package integration contains integration tests that run against a real Kubernetes cluster.
//
// These tests are excluded from normal `go test` runs and require the `integration` build tag.
// Run with: go test -tags=integration ./integration/...
//
// REUSABLE TEMPLATE:
// This file provides a reusable pattern for integration testing CLI tools.
// To adapt for another project:
// 1. Change binaryName to your project's binary name
// 2. Update buildBinary() with your build command
// 3. Modify cleanupResources() for your cleanup logic
// 4. Adjust healthCheck() for your service's health verification
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
	// binaryName is the name of the binary to test
	// TEMPLATE: Change this to your project's binary name
	binaryName = "kubeshark"

	// defaultTimeout for command execution
	defaultTimeout = 2 * time.Minute

	// startupTimeout for services that need time to initialize
	startupTimeout = 3 * time.Minute
)

var (
	// binaryPath caches the built binary path
	binaryPath string
	buildOnce  sync.Once
	buildErr   error
)

// requireKubernetesCluster skips the test if no Kubernetes cluster is available.
// TEMPLATE: Modify this check based on your infrastructure requirements.
func requireKubernetesCluster(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "cluster-info")
	if err := cmd.Run(); err != nil {
		t.Skip("Skipping: no Kubernetes cluster available (kubectl cluster-info failed)")
	}
}

// getKubesharkBinary returns the path to the kubeshark binary, building it if necessary.
// TEMPLATE: Adapt the build command for your project.
func getKubesharkBinary(t *testing.T) string {
	t.Helper()

	// Check if binary path is provided via environment
	if envBinary := os.Getenv("KUBESHARK_BINARY"); envBinary != "" {
		if _, err := os.Stat(envBinary); err == nil {
			return envBinary
		}
		t.Fatalf("KUBESHARK_BINARY set but file not found: %s", envBinary)
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

// buildBinary compiles the binary and returns its path.
// TEMPLATE: Change the build command and paths for your project.
func buildBinary(t *testing.T) (string, error) {
	t.Helper()

	// Find project root (directory containing go.mod)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", fmt.Errorf("finding project root: %w", err)
	}

	outputPath := filepath.Join(projectRoot, "bin", binaryName+"_integration_test")

	t.Logf("Building %s binary at %s", binaryName, outputPath)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// TEMPLATE: Modify this build command for your project
	cmd := exec.CommandContext(ctx, "go", "build",
		"-o", outputPath,
		filepath.Join(projectRoot, binaryName+".go"),
	)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %w\nOutput: %s", err, output)
	}

	return outputPath, nil
}

// findProjectRoot locates the project root by finding go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod in any parent directory")
		}
		dir = parent
	}
}

// runKubeshark executes the kubeshark binary with the given arguments.
// Returns combined stdout/stderr and any error.
func runKubeshark(t *testing.T, binary string, args ...string) (string, error) {
	t.Helper()
	return runKubesharkWithTimeout(t, binary, defaultTimeout, args...)
}

// runKubesharkWithTimeout executes the kubeshark binary with a custom timeout.
func runKubesharkWithTimeout(t *testing.T, binary string, timeout time.Duration, args ...string) (string, error) {
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
		return output, fmt.Errorf("command timed out after %v", timeout)
	}

	return output, err
}

// cleanupKubeshark ensures Kubeshark is not running in the cluster.
// TEMPLATE: Modify this for your project's cleanup requirements.
func cleanupKubeshark(t *testing.T, binary string) {
	t.Helper()

	if os.Getenv("INTEGRATION_SKIP_CLEANUP") == "true" {
		t.Log("Skipping cleanup (INTEGRATION_SKIP_CLEANUP=true)")
		return
	}

	t.Log("Cleaning up any existing Kubeshark installation...")

	// Run clean command, ignore errors (might not be installed)
	_, _ = runKubeshark(t, binary, "clean")

	// Wait a moment for resources to be deleted
	time.Sleep(2 * time.Second)
}

// waitForKubesharkReady waits for Kubeshark to be ready after starting.
// TEMPLATE: Modify this for your project's readiness check.
func waitForKubesharkReady(t *testing.T, binary string, timeout time.Duration) error {
	t.Helper()

	t.Log("Waiting for Kubeshark to be ready...")

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Check if pods are running
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-l", "app.kubernetes.io/name=kubeshark", "-o", "jsonpath={.items[*].status.phase}")
		output, err := cmd.Output()
		cancel()

		if err == nil && strings.Contains(string(output), "Running") {
			t.Log("Kubeshark is ready")
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for Kubeshark to be ready")
}

// isKubesharkRunning checks if Kubeshark is currently running in the cluster.
func isKubesharkRunning(t *testing.T) bool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-l", "app.kubernetes.io/name=kubeshark", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) != ""
}
