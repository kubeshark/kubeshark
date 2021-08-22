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

	tests := []int{50}

	for _, entriesCount := range tests {
		t.Run(fmt.Sprintf("%d", entriesCount), func(t *testing.T) {
			cliPath, cliPathErr := getCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			tapCmdArgs := getDefaultTapCommandArgs()
			tapCmd := exec.Command(cliPath, tapCmdArgs...)
			t.Logf("running command: %v", tapCmd.String())

			t.Cleanup(func() {
				if err := cleanupCommand(tapCmd); err != nil {
					t.Logf("failed to cleanup tap command, err: %v", err)
				}
			})

			if err := tapCmd.Start(); err != nil {
				t.Errorf("failed to start tap command, err: %v", err)
				return
			}

			if err := waitTapPodsReady(); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}

			proxyUrl := "http://localhost:8080/api/v1/namespaces/mizu-tests/services/httpbin/proxy/get"
			for i := 0; i < entriesCount; i++ {
				if _, requestErr := executeHttpRequest(proxyUrl); requestErr != nil {
					t.Errorf("failed to send proxy request, err: %v", requestErr)
					return
				}
			}

			entriesCheckFunc := func() error {
				timestamp := time.Now().UnixNano() / int64(time.Millisecond)

				entriesUrl := fmt.Sprintf("http://localhost:8899/mizu/api/entries?limit=%v&operator=lt&timestamp=%v", entriesCount, timestamp)
				requestResult, requestErr := executeHttpRequest(entriesUrl)
				if requestErr != nil {
					return fmt.Errorf("failed to get entries, err: %v", requestErr)
				}

				entries, ok := requestResult.([]interface{})
				if !ok {
					return fmt.Errorf("invalid entries type")
				}

				if len(entries) == 0 {
					return fmt.Errorf("unexpected entries result - Expected more than 0 entries")
				}

				entry, ok :=  entries[0].(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid entry type")
				}

				entryUrl := fmt.Sprintf("http://localhost:8899/mizu/api/entries/%v", entry["id"])
				requestResult, requestErr = executeHttpRequest(entryUrl)
				if requestErr != nil {
					return fmt.Errorf("failed to get entry, err: %v", requestErr)
				}

				if requestResult == nil {
					return fmt.Errorf("unexpected nil entry result")
				}

				return nil
			}
			if err := retriesExecute(ShortRetriesCount, entriesCheckFunc); err != nil {
				t.Errorf("%v", err)
				return
			}

			fetchCmdArgs := getDefaultFetchCommandArgs()
			fetchCmd := exec.Command(cliPath, fetchCmdArgs...)
			t.Logf("running command: %v", fetchCmd.String())

			if err := fetchCmd.Start(); err != nil {
				t.Errorf("failed to start fetch command, err: %v", err)
				return
			}

			harCheckFunc := func() error {
				harBytes, readFileErr := ioutil.ReadFile("./unknown_source.har")
				if readFileErr != nil {
					return fmt.Errorf("failed to read har file, err: %v", readFileErr)
				}

				harEntries, err := getEntriesFromHarBytes(harBytes)
				if err != nil {
					return fmt.Errorf("failed to get entries from har, err: %v", err)
				}

				if len(harEntries) == 0 {
					return fmt.Errorf("unexpected har entries result - Expected more than 0 entries")
				}

				return nil
			}
			if err := retriesExecute(ShortRetriesCount, harCheckFunc); err != nil {
				t.Errorf("%v", err)
				return
			}
		})
	}
}
