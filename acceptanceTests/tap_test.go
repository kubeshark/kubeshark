package acceptanceTests

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"
)

func TestTap(t *testing.T) {
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

			tapNamespace := getDefaultTapNamespace()
			tapCmdArgs = append(tapCmdArgs, tapNamespace...)

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

			apiServerUrl := getApiServerUrl(defaultApiServerPort)

			if err := waitTapPodsReady(apiServerUrl); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}

			proxyUrl := getProxyUrl(defaultNamespaceName, defaultServiceName)
			for i := 0; i < entriesCount; i++ {
				if _, requestErr := executeHttpGetRequest(fmt.Sprintf("%v/get", proxyUrl)); requestErr != nil {
					t.Errorf("failed to send proxy request, err: %v", requestErr)
					return
				}
			}

			entriesCheckFunc := func() error {
				timestamp := time.Now().UnixNano() / int64(time.Millisecond)

				entries, err := getDBEntries(timestamp, entriesCount, 1*time.Second)
				if err != nil {
					return err
				}
				err = checkEntriesAtLeast(entries, 1)
				if err != nil {
					return err
				}
				entry := entries[0]

				entryUrl := fmt.Sprintf("%v/entries/%v", apiServerUrl, entry["id"])
				requestResult, requestErr := executeHttpGetRequest(entryUrl)
				if requestErr != nil {
					return fmt.Errorf("failed to get entry, err: %v", requestErr)
				}

				if requestResult == nil {
					return fmt.Errorf("unexpected nil entry result")
				}

				return nil
			}
			if err := retriesExecute(shortRetriesCount, entriesCheckFunc); err != nil {
				t.Errorf("%v", err)
				return
			}
		})
	}
}

func TestTapGuiPort(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	tests := []uint16{8898}

	for _, guiPort := range tests {
		t.Run(fmt.Sprintf("%d", guiPort), func(t *testing.T) {
			cliPath, cliPathErr := getCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			tapCmdArgs := getDefaultTapCommandArgs()

			tapNamespace := getDefaultTapNamespace()
			tapCmdArgs = append(tapCmdArgs, tapNamespace...)

			tapCmdArgs = append(tapCmdArgs, "-p", fmt.Sprintf("%d", guiPort))

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

			apiServerUrl := getApiServerUrl(guiPort)

			if err := waitTapPodsReady(apiServerUrl); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}
		})
	}
}

func TestTapAllNamespaces(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	expectedPods := []PodDescriptor{
		{Name: "httpbin", Namespace: "mizu-tests"},
		{Name: "httpbin", Namespace: "mizu-tests2"},
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgs()
	tapCmdArgs = append(tapCmdArgs, "-A")

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

	apiServerUrl := getApiServerUrl(defaultApiServerPort)

	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	podsUrl := fmt.Sprintf("%v/status/tap", apiServerUrl)
	requestResult, requestErr := executeHttpGetRequest(podsUrl)
	if requestErr != nil {
		t.Errorf("failed to get tap status, err: %v", requestErr)
		return
	}

	pods, err := getPods(requestResult)
	if err != nil {
		t.Errorf("failed to get pods, err: %v", err)
		return
	}

	for _, expectedPod := range expectedPods {
		if !isPodDescriptorInPodArray(pods, expectedPod) {
			t.Errorf("unexpected result - expected pod not found, pod namespace: %v, pod name: %v", expectedPod.Namespace, expectedPod.Name)
			return
		}
	}
}

func TestTapMultipleNamespaces(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	expectedPods := []PodDescriptor{
		{Name: "httpbin", Namespace: "mizu-tests"},
		{Name: "httpbin2", Namespace: "mizu-tests"},
		{Name: "httpbin", Namespace: "mizu-tests2"},
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgs()
	var namespacesCmd []string
	for _, expectedPod := range expectedPods {
		namespacesCmd = append(namespacesCmd, "-n", expectedPod.Namespace)
	}
	tapCmdArgs = append(tapCmdArgs, namespacesCmd...)

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

	apiServerUrl := getApiServerUrl(defaultApiServerPort)

	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	podsUrl := fmt.Sprintf("%v/status/tap", apiServerUrl)
	requestResult, requestErr := executeHttpGetRequest(podsUrl)
	if requestErr != nil {
		t.Errorf("failed to get tap status, err: %v", requestErr)
		return
	}

	pods, err := getPods(requestResult)
	if err != nil {
		t.Errorf("failed to get pods, err: %v", err)
		return
	}

	if len(expectedPods) != len(pods) {
		t.Errorf("unexpected result - expected pods length: %v, actual pods length: %v", len(expectedPods), len(pods))
		return
	}

	for _, expectedPod := range expectedPods {
		if !isPodDescriptorInPodArray(pods, expectedPod) {
			t.Errorf("unexpected result - expected pod not found, pod namespace: %v, pod name: %v", expectedPod.Namespace, expectedPod.Name)
			return
		}
	}
}

func TestTapRegex(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	regexPodName := "httpbin2"
	expectedPods := []PodDescriptor{
		{Name: regexPodName, Namespace: "mizu-tests"},
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgsWithRegex(regexPodName)

	tapNamespace := getDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

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

	apiServerUrl := getApiServerUrl(defaultApiServerPort)

	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	podsUrl := fmt.Sprintf("%v/status/tap", apiServerUrl)
	requestResult, requestErr := executeHttpGetRequest(podsUrl)
	if requestErr != nil {
		t.Errorf("failed to get tap status, err: %v", requestErr)
		return
	}

	pods, err := getPods(requestResult)
	if err != nil {
		t.Errorf("failed to get pods, err: %v", err)
		return
	}

	if len(expectedPods) != len(pods) {
		t.Errorf("unexpected result - expected pods length: %v, actual pods length: %v", len(expectedPods), len(pods))
		return
	}

	for _, expectedPod := range expectedPods {
		if !isPodDescriptorInPodArray(pods, expectedPod) {
			t.Errorf("unexpected result - expected pod not found, pod namespace: %v, pod name: %v", expectedPod.Namespace, expectedPod.Name)
			return
		}
	}
}

func TestTapDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgs()

	tapNamespace := getDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	tapCmdArgs = append(tapCmdArgs, "--dry-run")

	tapCmd := exec.Command(cliPath, tapCmdArgs...)
	t.Logf("running command: %v", tapCmd.String())

	if err := tapCmd.Start(); err != nil {
		t.Errorf("failed to start tap command, err: %v", err)
		return
	}

	resultChannel := make(chan string, 1)

	go func() {
		if err := tapCmd.Wait(); err != nil {
			resultChannel <- "fail"
			return
		}
		resultChannel <- "success"
	}()

	go func() {
		time.Sleep(shortRetriesCount * time.Second)
		resultChannel <- "fail"
	}()

	testResult := <-resultChannel
	if testResult != "success" {
		t.Errorf("unexpected result - dry run cmd not done")
	}
}

func TestTapRedact(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgs()

	tapNamespace := getDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

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

	apiServerUrl := getApiServerUrl(defaultApiServerPort)

	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	proxyUrl := getProxyUrl(defaultNamespaceName, defaultServiceName)
	requestBody := map[string]string{"User": "Mizu"}
	for i := 0; i < defaultEntriesCount; i++ {
		if _, requestErr := executeHttpPostRequest(fmt.Sprintf("%v/post", proxyUrl), requestBody); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	redactCheckFunc := func() error {
		timestamp := time.Now().UnixNano() / int64(time.Millisecond)

		entries, err := getDBEntries(timestamp, defaultEntriesCount, 1*time.Second)
		if err != nil {
			return err
		}
		err = checkEntriesAtLeast(entries, 1)
		if err != nil {
			return err
		}
		firstEntry := entries[0]

		entryUrl := fmt.Sprintf("%v/entries/%v", apiServerUrl, firstEntry["id"])
		requestResult, requestErr := executeHttpGetRequest(entryUrl)
		if requestErr != nil {
			return fmt.Errorf("failed to get entry, err: %v", requestErr)
		}

		entry := requestResult.(map[string]interface{})["data"].(map[string]interface{})
		request := entry["request"].(map[string]interface{})

		headers := request["_headers"].([]interface{})
		for _, headerInterface := range headers {
			header := headerInterface.(map[string]interface{})
			if header["name"].(string) != "User-Agent" {
				continue
			}

			userAgent := header["value"].(string)
			if userAgent != "[REDACTED]" {
				return fmt.Errorf("unexpected result - user agent is not redacted")
			}
		}

		postData := request["postData"].(map[string]interface{})
		textDataStr := postData["text"].(string)

		var textData map[string]string
		if parseErr := json.Unmarshal([]byte(textDataStr), &textData); parseErr != nil {
			return fmt.Errorf("failed to parse text data, err: %v", parseErr)
		}

		if textData["User"] != "[REDACTED]" {
			return fmt.Errorf("unexpected result - user in body is not redacted")
		}

		return nil
	}
	if err := retriesExecute(shortRetriesCount, redactCheckFunc); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestTapNoRedact(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgs()

	tapNamespace := getDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	tapCmdArgs = append(tapCmdArgs, "--no-redact")

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

	apiServerUrl := getApiServerUrl(defaultApiServerPort)

	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	proxyUrl := getProxyUrl(defaultNamespaceName, defaultServiceName)
	requestBody := map[string]string{"User": "Mizu"}
	for i := 0; i < defaultEntriesCount; i++ {
		if _, requestErr := executeHttpPostRequest(fmt.Sprintf("%v/post", proxyUrl), requestBody); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	redactCheckFunc := func() error {
		timestamp := time.Now().UnixNano() / int64(time.Millisecond)

		entries, err := getDBEntries(timestamp, defaultEntriesCount, 1*time.Second)
		if err != nil {
			return err
		}
		err = checkEntriesAtLeast(entries, 1)
		if err != nil {
			return err
		}
		firstEntry := entries[0]

		entryUrl := fmt.Sprintf("%v/entries/%v", apiServerUrl, firstEntry["id"])
		requestResult, requestErr := executeHttpGetRequest(entryUrl)
		if requestErr != nil {
			return fmt.Errorf("failed to get entry, err: %v", requestErr)
		}

		entry := requestResult.(map[string]interface{})["data"].(map[string]interface{})
		request := entry["request"].(map[string]interface{})

		headers := request["_headers"].([]interface{})
		for _, headerInterface := range headers {
			header := headerInterface.(map[string]interface{})
			if header["name"].(string) != "User-Agent" {
				continue
			}

			userAgent := header["value"].(string)
			if userAgent == "[REDACTED]" {
				return fmt.Errorf("unexpected result - user agent is redacted")
			}
		}

		postData := request["postData"].(map[string]interface{})
		textDataStr := postData["text"].(string)

		var textData map[string]string
		if parseErr := json.Unmarshal([]byte(textDataStr), &textData); parseErr != nil {
			return fmt.Errorf("failed to parse text data, err: %v", parseErr)
		}

		if textData["User"] == "[REDACTED]" {
			return fmt.Errorf("unexpected result - user in body is redacted")
		}

		return nil
	}
	if err := retriesExecute(shortRetriesCount, redactCheckFunc); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestTapRegexMasking(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgs()

	tapNamespace := getDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	tapCmdArgs = append(tapCmdArgs, "-r", "Mizu")

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

	apiServerUrl := getApiServerUrl(defaultApiServerPort)

	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	proxyUrl := getProxyUrl(defaultNamespaceName, defaultServiceName)
	for i := 0; i < defaultEntriesCount; i++ {
		response, requestErr := http.Post(fmt.Sprintf("%v/post", proxyUrl), "text/plain", bytes.NewBufferString("Mizu"))
		if _, requestErr = executeHttpRequest(response, requestErr); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	redactCheckFunc := func() error {
		timestamp := time.Now().UnixNano() / int64(time.Millisecond)

		entries, err := getDBEntries(timestamp, defaultEntriesCount, 1*time.Second)
		if err != nil {
			return err
		}
		err = checkEntriesAtLeast(entries, 1)
		if err != nil {
			return err
		}
		firstEntry := entries[0]

		entryUrl := fmt.Sprintf("%v/entries/%v", apiServerUrl, firstEntry["id"])
		requestResult, requestErr := executeHttpGetRequest(entryUrl)
		if requestErr != nil {
			return fmt.Errorf("failed to get entry, err: %v", requestErr)
		}

		entry := requestResult.(map[string]interface{})["data"].(map[string]interface{})
		request := entry["request"].(map[string]interface{})

		postData := request["postData"].(map[string]interface{})
		textData := postData["text"].(string)

		if textData != "[REDACTED]" {
			return fmt.Errorf("unexpected result - body is not redacted")
		}

		return nil
	}
	if err := retriesExecute(shortRetriesCount, redactCheckFunc); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestTapIgnoredUserAgents(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgs()

	tapNamespace := getDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	ignoredUserAgentValue := "ignore"
	tapCmdArgs = append(tapCmdArgs, "--set", fmt.Sprintf("tap.ignored-user-agents=%v", ignoredUserAgentValue))

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

	apiServerUrl := getApiServerUrl(defaultApiServerPort)

	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	proxyUrl := getProxyUrl(defaultNamespaceName, defaultServiceName)

	ignoredUserAgentCustomHeader := "Ignored-User-Agent"
	headers := map[string]string{"User-Agent": ignoredUserAgentValue, ignoredUserAgentCustomHeader: ""}
	for i := 0; i < defaultEntriesCount; i++ {
		if _, requestErr := executeHttpGetRequestWithHeaders(fmt.Sprintf("%v/get", proxyUrl), headers); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < defaultEntriesCount; i++ {
		if _, requestErr := executeHttpGetRequest(fmt.Sprintf("%v/get", proxyUrl)); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	ignoredUserAgentsCheckFunc := func() error {
		timestamp := time.Now().UnixNano() / int64(time.Millisecond)

		entries, err := getDBEntries(timestamp, defaultEntriesCount, 1*time.Second)
		if err != nil {
			return err
		}
		err = checkEntriesAtLeast(entries, 1)
		if err != nil {
			return err
		}

		for _, entryInterface := range entries {
			entryUrl := fmt.Sprintf("%v/entries/%v", apiServerUrl, entryInterface["id"])
			requestResult, requestErr := executeHttpGetRequest(entryUrl)
			if requestErr != nil {
				return fmt.Errorf("failed to get entry, err: %v", requestErr)
			}

			entry := requestResult.(map[string]interface{})["data"].(map[string]interface{})
			request := entry["request"].(map[string]interface{})

			headers := request["_headers"].([]interface{})
			for _, headerInterface := range headers {
				header := headerInterface.(map[string]interface{})
				if header["name"].(string) != ignoredUserAgentCustomHeader {
					continue
				}

				return fmt.Errorf("unexpected result - user agent is not ignored")
			}
		}

		return nil
	}
	if err := retriesExecute(shortRetriesCount, ignoredUserAgentsCheckFunc); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestTapDumpLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgs()

	tapNamespace := getDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	tapCmdArgs = append(tapCmdArgs, "--set", "dump-logs=true")

	tapCmd := exec.Command(cliPath, tapCmdArgs...)
	t.Logf("running command: %v", tapCmd.String())

	if err := tapCmd.Start(); err != nil {
		t.Errorf("failed to start tap command, err: %v", err)
		return
	}

	apiServerUrl := getApiServerUrl(defaultApiServerPort)

	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	if err := cleanupCommand(tapCmd); err != nil {
		t.Errorf("failed to cleanup tap command, err: %v", err)
		return
	}

	mizuFolderPath, mizuPathErr := getMizuFolderPath()
	if mizuPathErr != nil {
		t.Errorf("failed to get mizu folder path, err: %v", mizuPathErr)
		return
	}

	files, readErr := ioutil.ReadDir(mizuFolderPath)
	if readErr != nil {
		t.Errorf("failed to read mizu folder files, err: %v", readErr)
		return
	}

	var dumpLogsPath string
	for _, file := range files {
		fileName := file.Name()
		if strings.Contains(fileName, "mizu_logs") {
			dumpLogsPath = path.Join(mizuFolderPath, fileName)
			break
		}
	}

	if dumpLogsPath == "" {
		t.Errorf("dump logs file not found")
		return
	}

	zipReader, zipError := zip.OpenReader(dumpLogsPath)
	if zipError != nil {
		t.Errorf("failed to get zip reader, err: %v", zipError)
		return
	}

	t.Cleanup(func() {
		if err := zipReader.Close(); err != nil {
			t.Logf("failed to close zip reader, err: %v", err)
		}
	})

	var logsFileNames []string
	for _, file := range zipReader.File {
		logsFileNames = append(logsFileNames, file.Name)
	}

	if !Contains(logsFileNames, "mizu.mizu-api-server.log") {
		t.Errorf("api server logs not found")
		return
	}

	if !Contains(logsFileNames, "mizu_cli.log") {
		t.Errorf("cli logs not found")
		return
	}

	if !Contains(logsFileNames, "mizu_events.log") {
		t.Errorf("events logs not found")
		return
	}

	if !ContainsPartOfValue(logsFileNames, "mizu.mizu-tapper-daemon-set") {
		t.Errorf("tapper logs not found")
		return
	}
}

func TestDaemonSeeTraffic(t *testing.T) {
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

			tapDaemonCmdArgs := getDefaultTapCommandArgsWithDaemonMode()

			tapNamespace := getDefaultTapNamespace()
			tapDaemonCmdArgs = append(tapDaemonCmdArgs, tapNamespace...)

			tapCmd := exec.Command(cliPath, tapDaemonCmdArgs...)

			viewCmd := exec.Command(cliPath, getDefaultViewCommandArgs()...)

			t.Cleanup(func() {
				daemonCleanup(t, viewCmd)
			})

			t.Logf("running command: %v", tapCmd.String())
			if err := tapCmd.Run(); err != nil {
				t.Errorf("error occured while running the tap command, err: %v", err)
				return
			}

			t.Logf("running command: %v", viewCmd.String())
			if err := viewCmd.Start(); err != nil {
				t.Errorf("error occured while running the view command, err: %v", err)
				return
			}

			apiServerUrl := getApiServerUrl(defaultApiServerPort)
			if err := waitTapPodsReady(apiServerUrl); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}

			proxyUrl := getProxyUrl(defaultNamespaceName, defaultServiceName)
			for i := 0; i < entriesCount; i++ {
				if _, requestErr := executeHttpGetRequest(fmt.Sprintf("%v/get", proxyUrl)); requestErr != nil {
					t.Errorf("failed to send proxy request, err: %v", requestErr)
					return
				}
			}

			entriesCheckFunc := func() error {
				timestamp := time.Now().UnixNano() / int64(time.Millisecond)

				entries, err := getDBEntries(timestamp, entriesCount, 1*time.Second)
				if err != nil {
					return err
				}
				err = checkEntriesAtLeast(entries, 1)
				if err != nil {
					return err
				}
				entry := entries[0]

				entryUrl := fmt.Sprintf("%v/entries/%v", apiServerUrl, entry["id"])
				requestResult, requestErr := executeHttpGetRequest(entryUrl)
				if requestErr != nil {
					return fmt.Errorf("failed to get entry, err: %v", requestErr)
				}

				if requestResult == nil {
					return fmt.Errorf("unexpected nil entry result")
				}

				return nil
			}
			if err := retriesExecute(shortRetriesCount, entriesCheckFunc); err != nil {
				t.Errorf("%v", err)
				return
			}
		})
	}
}

func TestDaemonMultipleNamespacesSeePods(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	expectedPods := []PodDescriptor{
		{Name: "httpbin", Namespace: "mizu-tests"},
		{Name: "httpbin2", Namespace: "mizu-tests"},
		{Name: "httpbin", Namespace: "mizu-tests2"},
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgsWithDaemonMode()
	var namespacesCmd []string
	for _, expectedPod := range expectedPods {
		namespacesCmd = append(namespacesCmd, "-n", expectedPod.Namespace)
	}
	tapCmdArgs = append(tapCmdArgs, namespacesCmd...)

	tapCmd := exec.Command(cliPath, tapCmdArgs...)

	viewCmd := exec.Command(cliPath, getDefaultViewCommandArgs()...)

	t.Cleanup(func() {
		daemonCleanup(t, viewCmd)
	})

	t.Logf("running command: %v", tapCmd.String())
	if err := tapCmd.Run(); err != nil {
		t.Errorf("failed to start tap command, err: %v", err)
		return
	}

	t.Logf("running command: %v", viewCmd.String())
	if err := viewCmd.Start(); err != nil {
		t.Errorf("error occured while running the view command, err: %v", err)
		return
	}

	apiServerUrl := getApiServerUrl(defaultApiServerPort)
	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	podsUrl := fmt.Sprintf("%v/status/tap", apiServerUrl)
	requestResult, requestErr := executeHttpGetRequest(podsUrl)
	if requestErr != nil {
		t.Errorf("failed to get tap status, err: %v", requestErr)
		return
	}

	pods, err := getPods(requestResult)
	if err != nil {
		t.Errorf("failed to get pods, err: %v", err)
		return
	}

	if len(expectedPods) != len(pods) {
		t.Errorf("unexpected result - expected pods length: %v, actual pods length: %v", len(expectedPods), len(pods))
		return
	}

	for _, expectedPod := range expectedPods {
		if !isPodDescriptorInPodArray(pods, expectedPod) {
			t.Errorf("unexpected result - expected pod not found, pod namespace: %v, pod name: %v", expectedPod.Namespace, expectedPod.Name)
			return
		}
	}
}

func TestDaemonSingleNamespaceSeePods(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	expectedPods := []PodDescriptor{
		{Name: "httpbin", Namespace: "mizu-tests"},
		{Name: "httpbin2", Namespace: "mizu-tests"},
	}
	unexpectedPods := []PodDescriptor{
		{Name: "httpbin", Namespace: "mizu-tests2"},
	}

	cliPath, cliPathErr := getCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := getDefaultTapCommandArgsWithDaemonMode()
	var namespacesCmd []string
	for _, expectedPod := range expectedPods {
		namespacesCmd = append(namespacesCmd, "-n", expectedPod.Namespace)
	}
	tapCmdArgs = append(tapCmdArgs, namespacesCmd...)

	tapCmd := exec.Command(cliPath, tapCmdArgs...)

	viewCmd := exec.Command(cliPath, getDefaultViewCommandArgs()...)

	t.Cleanup(func() {
		daemonCleanup(t, viewCmd)
	})

	t.Logf("running command: %v", tapCmd.String())
	if err := tapCmd.Run(); err != nil {
		t.Errorf("failed to start tap command, err: %v", err)
		return
	}

	t.Logf("running command: %v", viewCmd.String())
	if err := viewCmd.Start(); err != nil {
		t.Errorf("error occured while running the view command, err: %v", err)
		return
	}

	apiServerUrl := getApiServerUrl(defaultApiServerPort)
	if err := waitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	podsUrl := fmt.Sprintf("%v/status/tap", apiServerUrl)
	requestResult, requestErr := executeHttpGetRequest(podsUrl)
	if requestErr != nil {
		t.Errorf("failed to get tap status, err: %v", requestErr)
		return
	}

	pods, err := getPods(requestResult)
	if err != nil {
		t.Errorf("failed to get pods, err: %v", err)
		return
	}

	for _, unexpectedPod := range unexpectedPods {
		if isPodDescriptorInPodArray(pods, unexpectedPod) {
			t.Errorf("unexpected result - unexpected pod found, pod namespace: %v, pod name: %v", unexpectedPod.Namespace, unexpectedPod.Name)
			return
		}
	}

	if len(expectedPods) != len(pods) {
		t.Errorf("unexpected result - expected pods length: %v, actual pods length: %v", len(expectedPods), len(pods))
		return
	}

	for _, expectedPod := range expectedPods {
		if !isPodDescriptorInPodArray(pods, expectedPod) {
			t.Errorf("unexpected result - expected pod not found, pod namespace: %v, pod name: %v", expectedPod.Namespace, expectedPod.Name)
			return
		}
	}
}
