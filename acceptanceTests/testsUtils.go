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

	harLogInterface := har["log"]
	harLog, ok :=  harLogInterface.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid har log type")
	}

	harEntriesInterface := harLog["entries"]
	harEntries, ok := harEntriesInterface.([]interface{})
	if !ok {
		return nil, errors.New("invalid har entries type")
	}

	return harEntries, nil
}
