package acceptanceTests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/up9inc/mizu/shared"
)

const (
	longRetriesCount     = 100
	shortRetriesCount    = 10
	defaultApiServerPort = shared.DefaultApiServerPort
	defaultNamespaceName = "mizu-tests"
	defaultServiceName   = "httpbin"
	defaultEntriesCount  = 50
	waitAfterTapPodsReady = 3 * time.Second
	cleanCommandTimeout  = 1 * time.Minute
)

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
	return fmt.Sprintf("http://localhost:%v/mizu", port)
}

func getDefaultCommandArgs() []string {
	setFlag := "--set"
	telemetry := "telemetry=false"
	agentImage := "agent-image=gcr.io/up9-docker-hub/mizu/feature/tra-3842_daemon_mode1:0.0.0"
	imagePullPolicy := "image-pull-policy=Always"

	return []string{setFlag, telemetry, setFlag, agentImage, setFlag, imagePullPolicy}
}

func getDefaultTapCommandArgs() []string {
	tapCommand := "tap"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{tapCommand}, defaultCmdArgs...)
}

func getDefaultTapCommandArgsWithDaemonMode() []string {
	return append(getDefaultTapCommandArgs(), "--daemon")
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
	return []string{"clean"}
}

func getDefaultViewCommandArgs() []string {
	return []string{"view"}
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
	resolvingUrl := fmt.Sprintf("%v/status/tappersCount", apiServerUrl)
	tapPodsReadyFunc := func() error {
		requestResult, requestErr := executeHttpGetRequest(resolvingUrl)
		if requestErr != nil {
			return requestErr
		}

		tappersCount := requestResult.(float64)
		if tappersCount == 0 {
			return fmt.Errorf("no tappers running")
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

func executeHttpPostRequest(url string, body interface{}) (interface{}, error) {
	requestBody, jsonErr := json.Marshal(body)
	if jsonErr != nil {
		return nil, jsonErr
	}

	response, requestErr := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
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
	case err = <- commandDone:
		if err != nil {
			return err
		}
	case <- time.After(cleanCommandTimeout):
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
	tapStatus := tapStatusInterface.(map[string]interface{})
	podsInterface := tapStatus["pods"].([]interface{})

	var pods []map[string]interface{}
	for _, podInterface := range podsInterface {
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
