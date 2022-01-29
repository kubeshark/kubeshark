package acceptanceTests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/up9inc/mizu/shared"
)

const (
	longRetriesCount      = 100
	shortRetriesCount     = 10
	defaultApiServerPort  = shared.DefaultApiServerPort
	defaultNamespaceName  = "mizu-tests"
	defaultServiceName    = "httpbin"
	defaultEntriesCount   = 50
	waitAfterTapPodsReady = 3 * time.Second
	cleanCommandTimeout   = 1 * time.Minute
)

type PodDescriptor struct {
	Name      string
	Namespace string
}

func isPodDescriptorInPodArray(pods []map[string]interface{}, podDescriptor PodDescriptor) bool {
	for _, pod := range pods {
		podNamespace := pod["namespace"].(string)
		podName := pod["name"].(string)

		if podDescriptor.Namespace == podNamespace && strings.Contains(podName, podDescriptor.Name) {
			return true
		}
	}
	return false
}

func getCliPath() (string, error) {
	dir, filePathErr := os.Getwd()
	if filePathErr != nil {
		return "", filePathErr
	}

	cliPath := path.Join(dir, "../cli/bin/mizu_ci")
	return cliPath, nil
}

func getMizuFolderPath() (string, error) {
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return "", homeDirErr
	}

	return path.Join(home, ".mizu"), nil
}

func getConfigPath() (string, error) {
	mizuFolderPath, mizuPathError := getMizuFolderPath()
	if mizuPathError != nil {
		return "", mizuPathError
	}

	return path.Join(mizuFolderPath, "config.yaml"), nil
}

func getProxyUrl(namespace string, service string) string {
	return fmt.Sprintf("http://localhost:8080/api/v1/namespaces/%v/services/%v/proxy", namespace, service)
}

func getApiServerUrl(port uint16) string {
	return fmt.Sprintf("http://localhost:%v", port)
}

func getWebSocketUrl(port uint16) string {
	return fmt.Sprintf("ws://localhost:%v/ws", port)
}

func getDefaultCommandArgs() []string {
	setFlag := "--set"
	telemetry := "telemetry=false"
	agentImage := "agent-image=gcr.io/up9-docker-hub/mizu/ci:0.0.0"
	imagePullPolicy := "image-pull-policy=IfNotPresent"
	headless := "headless=true"

	return []string{setFlag, telemetry, setFlag, agentImage, setFlag, imagePullPolicy, setFlag, headless}
}

func getDefaultTapCommandArgs() []string {
	tapCommand := "tap"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{tapCommand}, defaultCmdArgs...)
}

func getDefaultTapCommandArgsWithRegex(regex string) []string {
	tapCommand := "tap"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{tapCommand, regex}, defaultCmdArgs...)
}

func getDefaultLogsCommandArgs() []string {
	logsCommand := "logs"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{logsCommand}, defaultCmdArgs...)
}

func getDefaultTapNamespace() []string {
	return []string{"-n", "mizu-tests"}
}

func getDefaultConfigCommandArgs() []string {
	configCommand := "config"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{configCommand}, defaultCmdArgs...)
}

func getDefaultCleanCommandArgs() []string {
	cleanCommand := "clean"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{cleanCommand}, defaultCmdArgs...)
}

func getDefaultViewCommandArgs() []string {
	viewCommand := "view"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{viewCommand}, defaultCmdArgs...)
}

func runCypressTests(t *testing.T, cypressRunCmd string) {
	cypressCmd := exec.Command("bash", "-c", cypressRunCmd)
	t.Logf("running command: %v", cypressCmd.String())
	out, err := cypressCmd.Output()
	if err != nil {
		t.Errorf("%s", out)
		return
	}
	t.Logf("%s", out)
}

func retriesExecute(retriesCount int, executeFunc func() error) error {
	var lastError interface{}

	for i := 0; i < retriesCount; i++ {
		if err := tryExecuteFunc(executeFunc); err != nil {
			lastError = err

			time.Sleep(1 * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("reached max retries count, retries count: %v, last err: %v", retriesCount, lastError)
}

func tryExecuteFunc(executeFunc func() error) (err interface{}) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = panicErr
		}
	}()

	return executeFunc()
}

func waitTapPodsReady(apiServerUrl string) error {
	resolvingUrl := fmt.Sprintf("%v/status/connectedTappersCount", apiServerUrl)
	tapPodsReadyFunc := func() error {
		requestResult, requestErr := executeHttpGetRequest(resolvingUrl)
		if requestErr != nil {
			return requestErr
		}

		connectedTappersCount := requestResult.(float64)
		if connectedTappersCount == 0 {
			return fmt.Errorf("no connected tappers running")
		}
		time.Sleep(waitAfterTapPodsReady)
		return nil
	}

	return retriesExecute(longRetriesCount, tapPodsReadyFunc)
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

func executeHttpGetRequestWithHeaders(url string, headers map[string]string) (interface{}, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for headerKey, headerValue := range headers {
		request.Header.Add(headerKey, headerValue)
	}

	client := &http.Client{}
	response, requestErr := client.Do(request)
	return executeHttpRequest(response, requestErr)
}

func executeHttpGetRequest(url string) (interface{}, error) {
	response, requestErr := http.Get(url)
	return executeHttpRequest(response, requestErr)
}

func executeHttpPostRequestWithHeaders(url string, headers map[string]string, body interface{}) (interface{}, error) {
	requestBody, jsonErr := json.Marshal(body)
	if jsonErr != nil {
		return nil, jsonErr
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	for headerKey, headerValue := range headers {
		request.Header.Add(headerKey, headerValue)
	}

	client := &http.Client{}
	response, requestErr := client.Do(request)
	return executeHttpRequest(response, requestErr)
}

func runMizuClean() error {
	cliPath, err := getCliPath()
	if err != nil {
		return err
	}

	cleanCmdArgs := getDefaultCleanCommandArgs()

	cleanCmd := exec.Command(cliPath, cleanCmdArgs...)

	commandDone := make(chan error)
	go func() {
		if err := cleanCmd.Run(); err != nil {
			commandDone <- err
		}
		commandDone <- nil
	}()

	select {
	case err = <-commandDone:
		if err != nil {
			return err
		}
	case <-time.After(cleanCommandTimeout):
		return errors.New("clean command timed out")
	}

	return nil
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

func getPods(tapStatusInterface interface{}) ([]map[string]interface{}, error) {
	tapPodsInterface := tapStatusInterface.([]interface{})

	var pods []map[string]interface{}
	for _, podInterface := range tapPodsInterface {
		pods = append(pods, podInterface.(map[string]interface{}))
	}

	return pods, nil
}

func getLogsPath() (string, error) {
	dir, filePathErr := os.Getwd()
	if filePathErr != nil {
		return "", filePathErr
	}

	logsPath := path.Join(dir, "mizu_logs.zip")
	return logsPath, nil
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	channel := make(chan struct{})
	go func() {
		defer close(channel)
		wg.Wait()
	}()
	select {
	case <-channel:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// checkEntriesAtLeast checks whether the number of entries greater than or equal to n
func checkEntriesAtLeast(entries []map[string]interface{}, n int) error {
	if len(entries) < n {
		return fmt.Errorf("Unexpected entries result - Expected more than %d entries", n-1)
	}
	return nil
}

// getDBEntries retrieves the entries from the database before the given timestamp.
// Also limits the results according to the limit parameter.
// Timeout for the WebSocket connection is defined by the timeout parameter.
func getDBEntries(timestamp int64, limit int, timeout time.Duration) (entries []map[string]interface{}, err error) {
	query := fmt.Sprintf("timestamp < %d and limit(%d)", timestamp, limit)
	webSocketUrl := getWebSocketUrl(defaultApiServerPort)

	var connection *websocket.Conn
	connection, _, err = websocket.DefaultDialer.Dial(webSocketUrl, nil)
	if err != nil {
		return
	}
	defer connection.Close()

	handleWSConnection := func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				return
			}

			var data map[string]interface{}
			if err = json.Unmarshal([]byte(message), &data); err != nil {
				return
			}

			if data["messageType"] == "entry" {
				entries = append(entries, data)
			}
		}
	}

	err = connection.WriteMessage(websocket.TextMessage, []byte(query))
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	go handleWSConnection(&wg)
	wg.Add(1)

	waitTimeout(&wg, timeout)

	return
}

func Contains(slice []string, containsValue string) bool {
	for _, sliceValue := range slice {
		if sliceValue == containsValue {
			return true
		}
	}

	return false
}

func ContainsPartOfValue(slice []string, containsValue string) bool {
	for _, sliceValue := range slice {
		if strings.Contains(sliceValue, containsValue) {
			return true
		}
	}

	return false
}

func RunCommandAsyncWithPipe(ctx context.Context, command string, args []string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	cmd := exec.CommandContext(ctx, command, args...)

	// Set output to Byte Buffers
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	err = cmd.Start()
	if err != nil {
		return stdout, stderr, fmt.Errorf("cmd.Start() failed: %s", err)
	}

	return stdout, stderr, nil
}
