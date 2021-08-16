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
	"testing"
	"time"
)

func TestSystemTapEntriesCount(t *testing.T) {
	tests := []int{1, 10, 20, 50 ,100}

	entriesCount := tests[4]

	dir, filePathErr := os.Getwd()
	if filePathErr != nil {
		t.Errorf("failed to get home dir, err: %v", filePathErr)
		return
	}

	mizuPath := path.Join(dir, "../cli/bin/mizu_ci")
	tapCommand := "tap"
	setFlag := "--set"
	namespaces := "tap.namespaces=mizu-tests"
	agentImage := "agent-image=gcr.io/up9-docker-hub/mizu/ci:0.0.0"
	imagePullPolicy := "image-pull-policy=Never"
	telemetry := "telemetry=false"

	cmd := exec.Command(mizuPath, tapCommand, setFlag, namespaces, setFlag, agentImage, setFlag, imagePullPolicy, setFlag, telemetry)

	t.Cleanup(func() {
		if err := cmd.Process.Signal(syscall.SIGQUIT); err != nil {
			t.Logf("failed to signal tap process, err: %v", err)
			return
		}

		if err := cmd.Wait(); err != nil {
			t.Logf("failed to wait tap process, err: %v", err)
			return
		}
	})

	if err := cmd.Start(); err != nil {
		t.Errorf("failed to start tap process, err: %v", err)
		return
	}

	time.Sleep(30 * time.Second)

	for i := 0; i < entriesCount; i++ {
		response, requestErr := http.Get("http://localhost:8080/api/v1/namespaces/mizu-tests/services/httpbin/proxy/get")
		if requestErr != nil {
			t.Errorf("failed to send request, err: %v", requestErr)
			return
		} else if response.StatusCode != 200 {
			t.Errorf("failed to send request, status code: %v", response.StatusCode)
			return
		}
	}

	time.Sleep(5 * time.Second)

	entriesUrl := fmt.Sprintf("http://localhost:8899/mizu/api/entries?limit=%v&operator=lt&timestamp=%v", entriesCount, time.Now().UnixNano())
	response, requestErr := http.Get(entriesUrl)
	if requestErr != nil {
		t.Errorf("failed to get entries, err: %v", requestErr)
		return
	} else if response.StatusCode != 200 {
		t.Errorf("failed to get entries, status code: %v", response.StatusCode)
		return
	}

	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		t.Errorf("failed to read entries, err: %v", readErr)
		return
	}

	var entries []interface{}
	if parseErr := json.Unmarshal(data, &entries); parseErr != nil {
		t.Errorf("failed to parse entries, err: %v", parseErr)
		return
	}

	if len(entries) != entriesCount {
		t.Errorf("unexpected result - Expected: %v, actual: %v", entriesCount, len(entries))
	}
}
