package integrationTests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/up9inc/mizu/shared"
)

const (
	AgentBin              = "../agent/build/mizuagent"
	InitializationTimeout = 5 * time.Second
	TestTimeout           = 60 * time.Second
	PCAPFile              = "http.cap"
)

func getApiServerUrl() string {
	return fmt.Sprintf("http://localhost:%v", shared.DefaultApiServerPort)
}

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
	ctx, _ := context.WithTimeout(context.Background(), TestTimeout)
	basenineCmd := exec.CommandContext(ctx, "basenine", "-port", shared.BaseninePort)
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
	for stay, timeout := true, time.After(InitializationTimeout); stay; {
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

func startAPIServer(t *testing.T, configPath string) (*exec.Cmd, io.ReadCloser, string) {
	args := []string{"--api-server"}
	if len(configPath) > 0 {
		args = append(args, "--config-path", configPath)
	}

	ctx, _ := context.WithTimeout(context.Background(), TestTimeout)
	apiServerCmd := exec.CommandContext(ctx, AgentBin, args...)
	t.Logf("running command: %v", apiServerCmd.String())

	t.Cleanup(func() {
		if err := cleanupCommand(apiServerCmd); err != nil {
			t.Logf("failed to cleanup API Server command, err: %v", err)
		}
	})

	out, err := apiServerCmd.StderrPipe()
	if err != nil {
		t.Errorf("failed to pipe API Server command output: %v", err)
		return nil, nil, ""
	}

	if err := apiServerCmd.Start(); err != nil {
		t.Errorf("failed to start API Server command: %v", err)
		return nil, nil, ""
	}

	// wait for some output
	buff := make([]byte, 32)
	for stay, timeout := true, time.After(InitializationTimeout); stay; {
		if n, err := out.Read(buff); err == nil {
			return apiServerCmd, out, string(buff[:n])
		}
		select {
		case <-timeout:
			stay = false
		default:
		}
	}

	t.Error("API Server command did not output any data in time")
	return nil, nil, ""
}

func startTapper(t *testing.T, pcapPath string) (*exec.Cmd, io.ReadCloser, string) {
	if len(pcapPath) == 0 {
		t.Error("tapper PCAP file path is empty")
		return nil, nil, ""
	}

	if !fileExists(pcapPath) {
		t.Errorf("tapper PCAP file does not exist: %s", pcapPath)
		return nil, nil, ""
	}

	if !strings.HasSuffix(pcapPath, ".cap") {
		t.Errorf("tapper PCAP file is not a valid .cap file: %s", pcapPath)
		return nil, nil, ""
	}

	args := []string{"-r", pcapPath, "--tap", "--api-server-address", "ws://localhost:8899/wsTapper"}
	ctx, _ := context.WithTimeout(context.Background(), TestTimeout)
	tapperCmd := exec.CommandContext(ctx, AgentBin, args...)
	t.Logf("running command: %v", tapperCmd.String())

	t.Cleanup(func() {
		if err := cleanupCommand(tapperCmd); err != nil {
			t.Logf("failed to cleanup tapper command, err: %v", err)
		}
	})

	out, err := tapperCmd.StderrPipe()
	if err != nil {
		t.Errorf("failed to pipe tapper command output: %v", err)
		return nil, nil, ""
	}

	if err := tapperCmd.Start(); err != nil {
		t.Errorf("failed to start tapper command: %v", err)
		return nil, nil, ""
	}

	// wait for some output
	buff := make([]byte, 32)
	for stay, timeout := true, time.After(InitializationTimeout); stay; {
		if n, err := out.Read(buff); err == nil {
			return tapperCmd, out, string(buff[:n])
		}
		select {
		case <-timeout:
			stay = false
		default:
		}
	}

	t.Error("tapper command did not output any data in time")
	return nil, nil, ""
}

func readOutput(output chan []byte, rc io.ReadCloser) {
	buff := make([]byte, 4096)
	for {
		n, err := rc.Read(buff)
		if err != nil {
			break
		}
		output <- buff[:n]
	}
}

func validateTapper(t *testing.T, wg *sync.WaitGroup, init string, rc io.ReadCloser) {
	wg.Add(1)

	tapperOutputChan := make(chan []byte)
	go readOutput(tapperOutputChan, rc)
	tapperOutput := <-tapperOutputChan
	rc.Close()

	defer wg.Done()

	output := fmt.Sprintf("%s%s", init, string(tapperOutput))
	t.Logf("Tapper output: %s\n", output)

	if !strings.Contains(output, "Starting tapper, websocket address: ws://localhost:8899/wsTapper") {
		t.Error("failed to validate tapper output")
		return
	}

	if !strings.Contains(output, fmt.Sprintf("Start reading packets from file-%s", PCAPFile)) {
		t.Error("failed to validate tapper output")
		return
	}

	if !strings.Contains(output, fmt.Sprintf("Got EOF while reading packets from file-%s", PCAPFile)) {
		t.Error("failed to validate tapper output")
		return
	}
}

func validateAPIServer(t *testing.T, wg *sync.WaitGroup, init string, rc io.ReadCloser) {
	wg.Add(1)

	apiServerOutputChan := make(chan []byte)
	go readOutput(apiServerOutputChan, rc)
	apiServerOutput := <-apiServerOutputChan
	rc.Close()

	defer wg.Done()

	output := fmt.Sprintf("%s%s", init, string(apiServerOutput))
	t.Logf("API Server output: %s\n", output)

	// validate extensions
	if !strings.Contains(output, "Initializing AMQP extension") {
		t.Error("failed to validate API Server AMQP extension initialization")
		return
	}
	if !strings.Contains(output, "Initializing HTTP extension") {
		t.Error("failed to validate API Server HTTP extension initialization")
		return
	}
	if !strings.Contains(output, "Initializing Kafka extension") {
		t.Error("failed to validate API Server Kafka extension initialization")
		return
	}
	if !strings.Contains(output, "Initializing Redis extension") {
		t.Error("failed to validate API Server Redis extension initialization")
		return
	}

	// server
	if !strings.Contains(output, "Starting the server") {
		t.Error("failed to validate API Server initialization")
		return
	}

	apiServerUrl := getApiServerUrl()
	requestResult, requestErr := executeHttpGetRequest(fmt.Sprintf("%v/status/connectedTappersCount", apiServerUrl))
	if requestErr != nil {
		t.Errorf("/status/connectedTappersCount request failed: %v", requestErr)
		return
	}
	connectedTappersCount := requestResult.(float64)
	if connectedTappersCount != 1 {
		t.Errorf("no connected tappers running - expected: 1, actual: %v", connectedTappersCount)
		return
	}

	requestResult, requestErr = executeHttpGetRequest(fmt.Sprintf("%v/status/general", apiServerUrl))
	if requestErr != nil {
		t.Errorf("/status/general request failed: %v", requestErr)
		return
	}
	generalStats := requestResult.(map[string]interface{})
	fmt.Println(generalStats)

}

func jsonBytesToInterface(jsonBytes []byte) (interface{}, error) {
	var result interface{}
	if parseErr := json.Unmarshal(jsonBytes, &result); parseErr != nil {
		return nil, parseErr
	}

	return result, nil
}

func executeHttpRequest(response *http.Response, requestErr error) (interface{}, error) {
	if requestErr != nil {
		return nil, requestErr
	} else if response.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code %v", response.StatusCode)
	}

	defer func() { response.Body.Close() }()

	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	return jsonBytesToInterface(data)
}

func executeHttpGetRequest(url string) (interface{}, error) {
	response, requestErr := http.Get(url)
	return executeHttpRequest(response, requestErr)
}

func Test(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	expectedBasenineOutput := fmt.Sprintf("Listening on :%s\n", shared.BaseninePort)
	expectedAgentOutput := "Initializing"

	_, basenineOutput := startBasenine(t)
	if !strings.HasSuffix(basenineOutput, expectedBasenineOutput) {
		t.Errorf("basenine is not running as expected - expected: %s, actual: %s", expectedBasenineOutput, basenineOutput)
	}

	_, apiServerReader, apiServerInit := startAPIServer(t, "")
	if !strings.HasSuffix(apiServerInit, expectedAgentOutput) {
		t.Errorf("API Server is not running as expected - expected: %s, actual: %s", expectedAgentOutput, apiServerInit)
	}

	_, tapperReader, tapperInit := startTapper(t, PCAPFile)
	if !strings.HasSuffix(tapperInit, expectedAgentOutput) {
		t.Errorf("Tapper is not running as expected - expected: %s, actual: %s", expectedAgentOutput, tapperInit)
	}

	// gives some time for api-server and tapper to initialize properly before validating the output
	for start := time.Now(); time.Since(start) < InitializationTimeout; {
		time.Sleep(1 * time.Second)
	}

	var wg = sync.WaitGroup{}

	validateTapper(t, &wg, tapperInit, tapperReader)
	validateAPIServer(t, &wg, apiServerInit, apiServerReader)

	wg.Wait()
}
