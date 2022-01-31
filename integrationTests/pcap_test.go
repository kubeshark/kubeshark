package integrationTests

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	AgentBin                = "../agent/build/mizuagent"
	BaseninePort            = "9099"
	BasenineCommandTimeout  = 5 * time.Second
	APIServerCommandTimeout = 30 * time.Second
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}

func cleanupCommand(cmd *exec.Cmd) error {
	if err := cmd.Process.Signal(syscall.SIGQUIT); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func startBasenine(t *testing.T) (*exec.Cmd, string) {
	basenineCmd := exec.Command("basenine", "-port", BaseninePort)
	t.Logf("running command: %v\n", basenineCmd.String())

	t.Cleanup(func() {
		if err := cleanupCommand(basenineCmd); err != nil {
			t.Logf("failed to cleanup basenine command, err: %v", err)
		}
	})

	// basenine outputs only to stderr
	out, err := basenineCmd.StderrPipe()
	if err != nil {
		t.Errorf("failed to pipe basenine command output: %v", err)
		return nil, ""
	}

	if err := basenineCmd.Start(); err != nil {
		t.Errorf("failed to start basenine command: %v", err)
		return nil, ""
	}

	// wait for some output
	buff := make([]byte, 64)
	for stay, timeout := true, time.After(BasenineCommandTimeout); stay; {
		if n, err := out.Read(buff); err == nil {
			return basenineCmd, string(buff[:n])
		}
		select {
		case <-timeout:
			stay = false
		default:
		}
	}

	t.Error("basenine command did not output any data in time")
	return nil, ""
}

func startAPIServer(t *testing.T, configPath string) (*exec.Cmd, string) {
	args := []string{"--api-server"}
	if len(configPath) > 0 {
		args = append(args, "--config-path", configPath)
	}
	apiServerCmd := exec.Command(AgentBin, args...)
	t.Logf("running command: %v", apiServerCmd.String())

	t.Cleanup(func() {
		if err := cleanupCommand(apiServerCmd); err != nil {
			t.Logf("failed to cleanup API Server command, err: %v", err)
		}
	})

	out, err := apiServerCmd.StderrPipe()
	if err != nil {
		t.Errorf("failed to pipe API Server command output: %v", err)
		return nil, ""
	}

	if err := apiServerCmd.Start(); err != nil {
		t.Errorf("failed to start API Server command: %v", err)
		return nil, ""
	}

	// wait for some output
	buff := make([]byte, 32)
	for stay, timeout := true, time.After(APIServerCommandTimeout); stay; {
		if n, err := out.Read(buff); err == nil {
			return apiServerCmd, string(buff[:n])
		}
		select {
		case <-timeout:
			stay = false
		default:
		}
	}

	t.Error("API Server command did not output any data in time")
	return nil, ""
}

func startTapper(t *testing.T, pcapPath string) (*exec.Cmd, string) {
	if len(pcapPath) == 0 {
		t.Error("tapper PCAP file path is empty")
		return nil, ""
	}

	if !fileExists(pcapPath) {
		t.Errorf("tapper PCAP file does not exist: %s", pcapPath)
		return nil, ""
	}

	if !strings.HasSuffix(pcapPath, ".cap") {
		t.Errorf("tapper PCAP file is not a valid .cap file: %s", pcapPath)
		return nil, ""
	}

	args := []string{"-r", pcapPath, "--tap", "--api-server-address ws://localhost:8899/wsTapper"}
	tapperCmd := exec.Command(AgentBin, args...)
	t.Logf("running command: %v", tapperCmd.String())

	t.Cleanup(func() {
		if err := cleanupCommand(tapperCmd); err != nil {
			t.Logf("failed to cleanup tapper command, err: %v", err)
		}
	})

	out, err := tapperCmd.StderrPipe()
	if err != nil {
		t.Errorf("failed to pipe tapper command output: %v", err)
		return nil, ""
	}

	if err := tapperCmd.Start(); err != nil {
		t.Errorf("failed to start tapper command: %v", err)
		return nil, ""
	}

	// wait for some output
	buff := make([]byte, 32)
	for stay, timeout := true, time.After(APIServerCommandTimeout); stay; {
		if n, err := out.Read(buff); err == nil {
			return tapperCmd, string(buff[:n])
		}
		select {
		case <-timeout:
			stay = false
		default:
		}
	}

	t.Error("tapper command did not output any data in time")
	return nil, ""
}

func Test(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}
	_, output := startBasenine(t)
	if !strings.HasSuffix(output, fmt.Sprintf("Listening on :%s\n", BaseninePort)) {
		t.Errorf("basenine is not running as expected: %s", output)
	}

	_, output = startAPIServer(t, "")
	if !strings.HasSuffix(output, "Initializing") {
		t.Errorf("API Server is not running as expected: %s", output)
	}

	_, output = startTapper(t, "http.cap")
	if !strings.HasSuffix(output, "Initializing") {
		t.Errorf("Tapper is not running as expected: %s", output)
	}
}
