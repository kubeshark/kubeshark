package acceptanceTests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"
)

const (
	longRetriesCount     = 100
	shortRetriesCount    = 10
	defaultApiServerPort = 8899
	defaultNamespaceName = "mizu-tests"
	defaultServiceName   = "httpbin"
	defaultEntriesCount  = 50
)

func getCliPath() (string, error) {
	dir, filePathErr := os.Getwd()
	if filePathErr != nil {
		return "", filePathErr
	}

	cliPath := path.Join(dir, "../cli/bin/mizu_ci")
	return cliPath, nil
}

func getConfigPath() (string, error) {
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return "", homeDirErr
	}

	return path.Join(home, ".mizu", "config.yaml"), nil
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
	agentImage := "agent-image=gcr.io/up9-docker-hub/mizu/ci:0.0.0"
	imagePullPolicy := "image-pull-policy=Never"

	return []string{setFlag, telemetry, setFlag, agentImage, setFlag, imagePullPolicy}
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

func getDefaultTapNamespace() []string {
	return []string{"-n", "mizu-tests"}
}

func getDefaultFetchCommandArgs() []string {
	fetchCommand := "fetch"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{fetchCommand}, defaultCmdArgs...)
}

func getDefaultConfigCommandArgs() []string {
	configCommand := "config"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{configCommand}, defaultCmdArgs...)
}

func retriesExecute(retriesCount int, executeFunc func() error) error {
	var lastError error

	for i := 0; i < retriesCount; i++ {
		if err := executeFunc(); err != nil {
			lastError = err

			time.Sleep(1 * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("reached max retries count, retries count: %v, last err: %v", retriesCount, lastError)
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

func cleanupCommand(cmd *exec.Cmd) error {
	if err := cmd.Process.Signal(syscall.SIGQUIT); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func getEntriesFromHarBytes(harBytes []byte) ([]interface{}, error) {
	harInterface, convertErr := jsonBytesToInterface(harBytes)
	if convertErr != nil {
		return nil, convertErr
	}

	har := harInterface.(map[string]interface{})
	harLog := har["log"].(map[string]interface{})
	harEntries := harLog["entries"].([]interface{})

	return harEntries, nil
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
