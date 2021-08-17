package tests

import (
	"encoding/json"
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

func GetDefaultTapCommandArgs() []string {
	tapCommand := "tap"
	setFlag := "--set"
	namespaces := "tap.namespaces=mizu-tests"
	agentImage := "agent-image=gcr.io/up9-docker-hub/mizu/ci:0.0.0"
	imagePullPolicy := "image-pull-policy=Never"
	telemetry := "telemetry=false"

	return []string{tapCommand, setFlag, namespaces, setFlag, agentImage, setFlag, imagePullPolicy, setFlag, telemetry}
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

	var result interface{}
	if parseErr := json.Unmarshal(data, &result); parseErr != nil {
		return nil, parseErr
	}

	return result, nil
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
