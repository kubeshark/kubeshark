package acceptanceTests

import (
	"encoding/json"
	"errors"
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
	LongRetriesCount = 100
	ShortRetriesCount = 5
)

func GetCliPath() (string, error) {
	dir, filePathErr := os.Getwd()
	if filePathErr != nil {
		return "", filePathErr
	}

	cliPath := path.Join(dir, "../cli/bin/mizu_ci")
	return cliPath, nil
}

func GetDefaultCommandArgs() []string {
	setFlag := "--set"
	telemetry := "telemetry=false"

	return []string{setFlag, telemetry}
}

func GetDefaultTapCommandArgs() []string {
	tapCommand := "tap"
	setFlag := "--set"
	namespaces := "tap.namespaces=mizu-tests"
	agentImage := "agent-image=gcr.io/up9-docker-hub/mizu/ci:0.0.0"
	imagePullPolicy := "image-pull-policy=Never"

	defaultCmdArgs := GetDefaultCommandArgs()

	return append([]string{tapCommand, setFlag, namespaces, setFlag, agentImage, setFlag, imagePullPolicy}, defaultCmdArgs...)
}

func GetDefaultFetchCommandArgs() []string {
	tapCommand := "fetch"

	defaultCmdArgs := GetDefaultCommandArgs()

	return append([]string{tapCommand}, defaultCmdArgs...)
}

func RetriesExecute(retriesCount int, executeFunc func() error) error {
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

func WaitTapPodsReady() error {
	resolvingUrl := fmt.Sprintf("http://localhost:8899/mizu/status/tappersCount")
	tapPodsReadyFunc := func() error {
		requestResult, requestErr := ExecuteHttpRequest(resolvingUrl)
		if requestErr != nil {
			return requestErr
		}

		tappersCount, ok := requestResult.(float64)
		if !ok {
			return fmt.Errorf("invalid tappers count type")
		}

		if tappersCount == 0 {
			return fmt.Errorf("no tappers running")
		}

		return nil
	}

	return RetriesExecute(LongRetriesCount, tapPodsReadyFunc)
}

func JsonBytesToInterface(jsonBytes []byte) (interface{}, error) {
	var result interface{}
	if parseErr := json.Unmarshal(jsonBytes, &result); parseErr != nil {
		return nil, parseErr
	}

	return result, nil
}

func ExecuteHttpRequest(url string) (interface{}, error) {
	response, requestErr := http.Get(url)
	if requestErr != nil {
		return nil, requestErr
	} else if response.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code %v", response.StatusCode)
	}

	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	return JsonBytesToInterface(data)
}

func CleanupCommand(cmd *exec.Cmd) error {
	if err := cmd.Process.Signal(syscall.SIGQUIT); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func GetEntriesFromHarBytes(harBytes []byte) ([]interface{}, error){
	harInterface, convertErr := JsonBytesToInterface(harBytes)
	if convertErr != nil {
		return nil, convertErr
	}

	har, ok := harInterface.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid har type")
	}

	harLog, ok :=  har["log"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid har log type")
	}

	harEntries, ok := harLog["entries"].([]interface{})
	if !ok {
		return nil, errors.New("invalid har entries type")
	}

	return harEntries, nil
}
