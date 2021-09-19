package acceptanceTests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
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

				entriesUrl := fmt.Sprintf("%v/api/entries?limit=%v&operator=lt&timestamp=%v", apiServerUrl, entriesCount, timestamp)
				requestResult, requestErr := executeHttpGetRequest(entriesUrl)
				if requestErr != nil {
					return fmt.Errorf("failed to get entries, err: %v", requestErr)
				}

				entries := requestResult.([]interface{})
				if len(entries) == 0 {
					return fmt.Errorf("unexpected entries result - Expected more than 0 entries")
				}

				entry :=  entries[0].(map[string]interface{})

				entryUrl := fmt.Sprintf("%v/api/entries/%v", apiServerUrl, entry["id"])
				requestResult, requestErr = executeHttpGetRequest(entryUrl)
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

	expectedPods := []struct{
		Name      string
		Namespace string
	}{
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

	podsUrl := fmt.Sprintf("%v/api/tapStatus", apiServerUrl)
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
		podFound := false

		for _, pod := range pods {
			podNamespace :=  pod["namespace"].(string)
			podName := pod["name"].(string)

			if expectedPod.Namespace == podNamespace && strings.Contains(podName, expectedPod.Name) {
				podFound = true
				break
			}
		}

		if !podFound {
			t.Errorf("unexpected result - expected pod not found, pod namespace: %v, pod name: %v", expectedPod.Namespace, expectedPod.Name)
			return
		}
	}
}

func TestTapMultipleNamespaces(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	expectedPods := []struct{
		Name      string
		Namespace string
	}{
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

	podsUrl := fmt.Sprintf("%v/api/tapStatus", apiServerUrl)
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
		podFound := false

		for _, pod := range pods {
			podNamespace :=  pod["namespace"].(string)
			podName := pod["name"].(string)

			if expectedPod.Namespace == podNamespace && strings.Contains(podName, expectedPod.Name) {
				podFound = true
				break
			}
		}

		if !podFound {
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
	expectedPods := []struct{
		Name      string
		Namespace string
	}{
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

	podsUrl := fmt.Sprintf("%v/api/tapStatus", apiServerUrl)
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
		podFound := false

		for _, pod := range pods {
			podNamespace :=  pod["namespace"].(string)
			podName := pod["name"].(string)

			if expectedPod.Namespace == podNamespace && strings.Contains(podName, expectedPod.Name) {
				podFound = true
				break
			}
		}

		if !podFound {
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

	testResult := <- resultChannel
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

		entriesUrl := fmt.Sprintf("%v/api/entries?limit=%v&operator=lt&timestamp=%v", apiServerUrl, defaultEntriesCount, timestamp)
		requestResult, requestErr := executeHttpGetRequest(entriesUrl)
		if requestErr != nil {
			return fmt.Errorf("failed to get entries, err: %v", requestErr)
		}

		entries := requestResult.([]interface{})
		if len(entries) == 0 {
			return fmt.Errorf("unexpected entries result - Expected more than 0 entries")
		}

		firstEntry :=  entries[0].(map[string]interface{})

		entryUrl := fmt.Sprintf("%v/api/entries/%v", apiServerUrl, firstEntry["id"])
		requestResult, requestErr = executeHttpGetRequest(entryUrl)
		if requestErr != nil {
			return fmt.Errorf("failed to get entry, err: %v", requestErr)
		}

		data := requestResult.(map[string]interface{})["data"].(map[string]interface{})
		entryJson := data["entry"].(string)

		var entry map[string]interface{}
		if parseErr := json.Unmarshal([]byte(entryJson), &entry); parseErr != nil {
			return fmt.Errorf("failed to parse entry, err: %v", parseErr)
		}

		entryRequest := entry["request"].(map[string]interface{})
		entryPayload := entryRequest["payload"].(map[string]interface{})
		entryDetails := entryPayload["details"].(map[string]interface{})

		headers :=  entryDetails["headers"].([]interface{})
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

		postData := entryDetails["postData"].(map[string]interface{})
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

		entriesUrl := fmt.Sprintf("%v/api/entries?limit=%v&operator=lt&timestamp=%v", apiServerUrl, defaultEntriesCount, timestamp)
		requestResult, requestErr := executeHttpGetRequest(entriesUrl)
		if requestErr != nil {
			return fmt.Errorf("failed to get entries, err: %v", requestErr)
		}

		entries := requestResult.([]interface{})
		if len(entries) == 0 {
			return fmt.Errorf("unexpected entries result - Expected more than 0 entries")
		}

		firstEntry :=  entries[0].(map[string]interface{})

		entryUrl := fmt.Sprintf("%v/api/entries/%v", apiServerUrl, firstEntry["id"])
		requestResult, requestErr = executeHttpGetRequest(entryUrl)
		if requestErr != nil {
			return fmt.Errorf("failed to get entry, err: %v", requestErr)
		}

		data := requestResult.(map[string]interface{})["data"].(map[string]interface{})
		entryJson := data["entry"].(string)

		var entry map[string]interface{}
		if parseErr := json.Unmarshal([]byte(entryJson), &entry); parseErr != nil {
			return fmt.Errorf("failed to parse entry, err: %v", parseErr)
		}

		entryRequest := entry["request"].(map[string]interface{})
		entryPayload := entryRequest["payload"].(map[string]interface{})
		entryDetails := entryPayload["details"].(map[string]interface{})

		headers :=  entryDetails["headers"].([]interface{})
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

		postData := entryDetails["postData"].(map[string]interface{})
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

		entriesUrl := fmt.Sprintf("%v/api/entries?limit=%v&operator=lt&timestamp=%v", apiServerUrl, defaultEntriesCount, timestamp)
		requestResult, requestErr := executeHttpGetRequest(entriesUrl)
		if requestErr != nil {
			return fmt.Errorf("failed to get entries, err: %v", requestErr)
		}

		entries := requestResult.([]interface{})
		if len(entries) == 0 {
			return fmt.Errorf("unexpected entries result - Expected more than 0 entries")
		}

		firstEntry :=  entries[0].(map[string]interface{})

		entryUrl := fmt.Sprintf("%v/api/entries/%v", apiServerUrl, firstEntry["id"])
		requestResult, requestErr = executeHttpGetRequest(entryUrl)
		if requestErr != nil {
			return fmt.Errorf("failed to get entry, err: %v", requestErr)
		}

		data := requestResult.(map[string]interface{})["data"].(map[string]interface{})
		entryJson := data["entry"].(string)

		var entry map[string]interface{}
		if parseErr := json.Unmarshal([]byte(entryJson), &entry); parseErr != nil {
			return fmt.Errorf("failed to parse entry, err: %v", parseErr)
		}

		entryRequest := entry["request"].(map[string]interface{})
		entryPayload := entryRequest["payload"].(map[string]interface{})
		entryDetails := entryPayload["details"].(map[string]interface{})

		postData := entryDetails["postData"].(map[string]interface{})
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
