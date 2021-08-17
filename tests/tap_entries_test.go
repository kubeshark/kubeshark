package tests

import (
	"fmt"
	"os/exec"
	"testing"
	"time"
)

func TestSystemTapEntriesCount(t *testing.T) {
	tests := []int{1, 10, 20, 50, 100}

	for _, entriesCount := range tests {
		t.Run(fmt.Sprintf("%d", entriesCount), func(t *testing.T) {
			cliPath, cliPathErr := GetCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			tapCmdArgs := GetDefaultTapCommandArgs()
			cmd := exec.Command(cliPath, tapCmdArgs...)
			t.Logf("running command: %v", cmd.String())

			t.Cleanup(func() {
				if err := CleanupCommand(cmd); err != nil {
					t.Logf("failed to cleanup command, err: %v", err)
				}
			})

			if err := cmd.Start(); err != nil {
				t.Errorf("failed to start tap process, err: %v", err)
				return
			}

			time.Sleep(30 * time.Second)

			proxyUrl := "http://localhost:8080/api/v1/namespaces/mizu-tests/services/httpbin/proxy/get"
			for i := 0; i < entriesCount; i++ {
				if _, requestErr := ExecuteHttpRequest(proxyUrl); requestErr != nil {
					t.Errorf("failed to send proxy request, err: %v", requestErr)
					return
				}
			}

			time.Sleep(5 * time.Second)

			entriesUrl := fmt.Sprintf("http://localhost:8899/mizu/api/entries?limit=%v&operator=lt&timestamp=%v", entriesCount, time.Now().UnixNano())
			requestResult, requestErr := ExecuteHttpRequest(entriesUrl)
			if requestErr != nil {
				t.Errorf("failed to get entries, err: %v", requestErr)
				return
			}

			entries, ok := requestResult.([]interface{})
			if !ok {
				t.Errorf("invalid entries type")
				return
			}

			if len(entries) != entriesCount {
				t.Errorf("unexpected result - Expected: %v, actual: %v", entriesCount, len(entries))
			}
		})
	}
}
