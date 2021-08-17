package acceptanceTests

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"
	"time"
)

func TestTapAndFetch(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	tests := []int{1, 100}

	for _, entriesCount := range tests {
		t.Run(fmt.Sprintf("%d", entriesCount), func(t *testing.T) {
			cliPath, cliPathErr := GetCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			tapCmdArgs := GetDefaultTapCommandArgs()
			tapCmd := exec.Command(cliPath, tapCmdArgs...)
			t.Logf("running command: %v", tapCmd.String())

			t.Cleanup(func() {
				if err := CleanupCommand(tapCmd); err != nil {
					t.Logf("failed to cleanup tap command, err: %v", err)
				}
			})

			if err := tapCmd.Start(); err != nil {
				t.Errorf("failed to start tap command, err: %v", err)
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
			timestamp := time.Now().UnixNano() / int64(time.Millisecond)

			entriesUrl := fmt.Sprintf("http://localhost:8899/mizu/api/entries?limit=%v&operator=lt&timestamp=%v", entriesCount, timestamp)
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
				t.Errorf("unexpected entries result - Expected: %v, actual: %v", entriesCount, len(entries))
				return
			}

			entry, ok :=  entries[0].(map[string]interface{})
			if !ok {
				t.Errorf("invalid entry type")
				return
			}

			entryUrl := fmt.Sprintf("http://localhost:8899/mizu/api/entries/%v", entry["id"])
			requestResult, requestErr = ExecuteHttpRequest(entryUrl)
			if requestErr != nil {
				t.Errorf("failed to get entry, err: %v", requestErr)
				return
			}

			if requestResult == nil {
				t.Errorf("unexpected nil entry result")
				return
			}

			fetchCmdArgs := GetDefaultFetchCommandArgs()
			fetchCmd := exec.Command(cliPath, fetchCmdArgs...)
			t.Logf("running command: %v", fetchCmd.String())

			t.Cleanup(func() {
				if err := CleanupCommand(fetchCmd); err != nil {
					t.Logf("failed to cleanup fetch command, err: %v", err)
				}
			})

			if err := fetchCmd.Start(); err != nil {
				t.Errorf("failed to start fetch command, err: %v", err)
				return
			}

			time.Sleep(5 * time.Second)

			harBytes, readFileErr := ioutil.ReadFile("./unknown_source.har")
			if readFileErr != nil {
				t.Errorf("failed to read har file, err: %v", readFileErr)
				return
			}

			harEntries, err := GetEntriesFromHarBytes(harBytes)
			if err != nil {
				t.Errorf("failed to get entries from har, err: %v", err)
				return
			}

			if len(harEntries) != entriesCount {
				t.Errorf("unexpected har entries result - Expected: %v, actual: %v", entriesCount, len(harEntries))
				return
			}
		})
	}
}
