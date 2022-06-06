package acceptanceTests

import (
	"archive/zip"
	"bytes"
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
	basicTapTest(t, false)
}

func basicTapTest(t *testing.T, shouldCheckSrcAndDest bool, extraArgs... string) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	tests := []int{50}

	for _, entriesCount := range tests {
		t.Run(fmt.Sprintf("%d", entriesCount), func(t *testing.T) {
			cliPath, cliPathErr := GetCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			tapCmdArgs := GetDefaultTapCommandArgs()

			tapNamespace := GetDefaultTapNamespace()
			tapCmdArgs = append(tapCmdArgs, tapNamespace...)

			tapCmdArgs = append(tapCmdArgs, extraArgs...)

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

			apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

			if err := WaitTapPodsReady(apiServerUrl); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}

			proxyUrl := GetProxyUrl(DefaultNamespaceName, DefaultServiceName)
			for i := 0; i < entriesCount; i++ {
				if _, requestErr := ExecuteHttpGetRequest(fmt.Sprintf("%v/get", proxyUrl)); requestErr != nil {
					t.Errorf("failed to send proxy request, err: %v", requestErr)
					return
				}
			}

			expectedPods := []PodDescriptor{
				{Name: "httpbin", Namespace: "mizu-tests"},
				{Name: "httpbin2", Namespace: "mizu-tests"},
			}

			var expectedPodsStr string
			for i := 0; i < len(expectedPods); i++ {
				expectedPodsStr += fmt.Sprintf("Name:%vNamespace:%v", expectedPods[i].Name, expectedPods[i].Namespace)
			}

			RunCypressTests(t, fmt.Sprintf("npx cypress@9.5.4 run --spec  \"cypress/integration/tests/UiTest.js\" --env entriesCount=%d,arrayDict=%v,shouldCheckSrcAndDest=%v",
				entriesCount, expectedPodsStr, shouldCheckSrcAndDest))
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
			cliPath, cliPathErr := GetCliPath()
			if cliPathErr != nil {
				t.Errorf("failed to get cli path, err: %v", cliPathErr)
				return
			}

			tapCmdArgs := GetDefaultTapCommandArgs()

			tapNamespace := GetDefaultTapNamespace()
			tapCmdArgs = append(tapCmdArgs, tapNamespace...)

			tapCmdArgs = append(tapCmdArgs, "-p", fmt.Sprintf("%d", guiPort))

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

			apiServerUrl := GetApiServerUrl(guiPort)

			if err := WaitTapPodsReady(apiServerUrl); err != nil {
				t.Errorf("failed to start tap pods on time, err: %v", err)
				return
			}

			proxyUrl := GetProxyUrl(DefaultNamespaceName, DefaultServiceName)
			for i := 0; i < DefaultEntriesCount; i++ {
				if _, requestErr := ExecuteHttpGetRequest(fmt.Sprintf("%v/get", proxyUrl)); requestErr != nil {
					t.Errorf("failed to send proxy request, err: %v", requestErr)
					return
				}
			}

			RunCypressTests(t, fmt.Sprintf("npx cypress@9.5.4 run --spec \"cypress/integration/tests/GuiPort.js\" --env name=%v,namespace=%v,port=%d",
				"httpbin", "mizu-tests", guiPort))
		})
	}
}

func TestTapAllNamespaces(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	expectedPods := []PodDescriptor{
		{Name: "httpbin", Namespace: "mizu-tests"},
		{Name: "httpbin2", Namespace: "mizu-tests"},
		{Name: "httpbin", Namespace: "mizu-tests2"},
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()
	tapCmdArgs = append(tapCmdArgs, "-A")

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

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	RunCypressTests(t, fmt.Sprintf("npx cypress@9.5.4 run --spec  \"cypress/integration/tests/MultipleNamespaces.js\" --env name1=%v,name2=%v,name3=%v,namespace1=%v,namespace2=%v,namespace3=%v",
		expectedPods[0].Name, expectedPods[1].Name, expectedPods[2].Name, expectedPods[0].Namespace, expectedPods[1].Namespace, expectedPods[2].Namespace))
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

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()
	var namespacesCmd []string
	for _, expectedPod := range expectedPods {
		namespacesCmd = append(namespacesCmd, "-n", expectedPod.Namespace)
	}
	tapCmdArgs = append(tapCmdArgs, namespacesCmd...)

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

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	RunCypressTests(t, fmt.Sprintf("npx cypress@9.5.4 run --spec  \"cypress/integration/tests/MultipleNamespaces.js\" --env name1=%v,name2=%v,name3=%v,namespace1=%v,namespace2=%v,namespace3=%v",
		expectedPods[0].Name, expectedPods[1].Name, expectedPods[2].Name, expectedPods[0].Namespace, expectedPods[1].Namespace, expectedPods[2].Namespace))
}

func TestTapRegex(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	regexPodName := "httpbin2"
	expectedPods := []PodDescriptor{
		{Name: regexPodName, Namespace: "mizu-tests"},
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgsWithRegex(regexPodName)

	tapNamespace := GetDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

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

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	RunCypressTests(t, fmt.Sprintf("npx cypress@9.5.4 run --spec  \"cypress/integration/tests/Regex.js\" --env name=%v,namespace=%v",
		expectedPods[0].Name, expectedPods[0].Namespace))
}

func TestTapDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()

	tapNamespace := GetDefaultTapNamespace()
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
		time.Sleep(ShortRetriesCount * time.Second)
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

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()

	tapNamespace := GetDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)
	tapCmdArgs = append(tapCmdArgs, "--redact")

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

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	proxyUrl := GetProxyUrl(DefaultNamespaceName, DefaultServiceName)
	requestHeaders := map[string]string{"User-Header": "Mizu"}
	requestBody := map[string]string{"User": "Mizu"}
	for i := 0; i < DefaultEntriesCount; i++ {
		if _, requestErr := ExecuteHttpPostRequestWithHeaders(fmt.Sprintf("%v/post", proxyUrl), requestHeaders, requestBody); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	RunCypressTests(t, "npx cypress@9.5.4 run --spec  \"cypress/integration/tests/Redact.js\"")
}

func TestTapNoRedact(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()

	tapNamespace := GetDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

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

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	proxyUrl := GetProxyUrl(DefaultNamespaceName, DefaultServiceName)
	requestHeaders := map[string]string{"User-Header": "Mizu"}
	requestBody := map[string]string{"User": "Mizu"}
	for i := 0; i < DefaultEntriesCount; i++ {
		if _, requestErr := ExecuteHttpPostRequestWithHeaders(fmt.Sprintf("%v/post", proxyUrl), requestHeaders, requestBody); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	RunCypressTests(t, "npx cypress@9.5.4 run --spec  \"cypress/integration/tests/NoRedact.js\"")
}

func TestTapRegexMasking(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()

	tapNamespace := GetDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	tapCmdArgs = append(tapCmdArgs, "--redact")

	tapCmdArgs = append(tapCmdArgs, "-r", "Mizu")

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

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	proxyUrl := GetProxyUrl(DefaultNamespaceName, DefaultServiceName)
	for i := 0; i < DefaultEntriesCount; i++ {
		response, requestErr := http.Post(fmt.Sprintf("%v/post", proxyUrl), "text/plain", bytes.NewBufferString("Mizu"))
		if _, requestErr = ExecuteHttpRequest(response, requestErr); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	RunCypressTests(t, "npx cypress@9.5.4 run --spec \"cypress/integration/tests/RegexMasking.js\"")

}

func TestTapIgnoredUserAgents(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()

	tapNamespace := GetDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	ignoredUserAgentValue := "ignore"
	tapCmdArgs = append(tapCmdArgs, "--set", fmt.Sprintf("tap.ignored-user-agents=%v", ignoredUserAgentValue))

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

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	proxyUrl := GetProxyUrl(DefaultNamespaceName, DefaultServiceName)

	ignoredUserAgentCustomHeader := "Ignored-User-Agent"
	headers := map[string]string{"User-Agent": ignoredUserAgentValue, ignoredUserAgentCustomHeader: ""}
	for i := 0; i < DefaultEntriesCount; i++ {
		if _, requestErr := ExecuteHttpGetRequestWithHeaders(fmt.Sprintf("%v/get", proxyUrl), headers); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	for i := 0; i < DefaultEntriesCount; i++ {
		if _, requestErr := ExecuteHttpGetRequest(fmt.Sprintf("%v/get", proxyUrl)); requestErr != nil {
			t.Errorf("failed to send proxy request, err: %v", requestErr)
			return
		}
	}

	RunCypressTests(t, "npx cypress@9.5.4 run --spec  \"cypress/integration/tests/IgnoredUserAgents.js\"")
}

func TestTapDumpLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("ignored acceptance test")
	}

	cliPath, cliPathErr := GetCliPath()
	if cliPathErr != nil {
		t.Errorf("failed to get cli path, err: %v", cliPathErr)
		return
	}

	tapCmdArgs := GetDefaultTapCommandArgs()

	tapNamespace := GetDefaultTapNamespace()
	tapCmdArgs = append(tapCmdArgs, tapNamespace...)

	tapCmdArgs = append(tapCmdArgs, "--set", "dump-logs=true")

	tapCmd := exec.Command(cliPath, tapCmdArgs...)
	t.Logf("running command: %v", tapCmd.String())

	if err := tapCmd.Start(); err != nil {
		t.Errorf("failed to start tap command, err: %v", err)
		return
	}

	apiServerUrl := GetApiServerUrl(DefaultApiServerPort)

	if err := WaitTapPodsReady(apiServerUrl); err != nil {
		t.Errorf("failed to start tap pods on time, err: %v", err)
		return
	}

	if err := CleanupCommand(tapCmd); err != nil {
		t.Errorf("failed to cleanup tap command, err: %v", err)
		return
	}

	mizuFolderPath, mizuPathErr := GetMizuFolderPath()
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

	if !Contains(logsFileNames, "mizu.mizu-api-server.mizu-api-server.log") {
		t.Errorf("api server logs not found")
		return
	}

	if !Contains(logsFileNames, "mizu.mizu-api-server.basenine.log") {
		t.Errorf("basenine logs not found")
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

func TestIpResolving(t *testing.T) {
	namespace := AllNamespaces

	t.Log("add permissions for ip-resolution for current user")
	if err := ApplyKubeFilesForTest(
		t,
		"minikube",
		namespace,
		"../cli/cmd/permissionFiles/permissions-all-namespaces-ip-resolution-optional.yaml",
	); err != nil {
		t.Errorf("failed to create k8s permissions, %v", err)
		return
	}

	basicTapTest(t, true)
}

func TestRestrictedMode(t *testing.T) {
	namespace := "mizu-tests"

	t.Log("creating permissions for restricted user")
	if err := ApplyKubeFilesForTest(
		t,
		"minikube",
		namespace,
		"../cli/cmd/permissionFiles/permissions-ns-tap.yaml",
	); err != nil {
		t.Errorf("failed to create k8s permissions, %v", err)
		return
	}

	t.Log("switching k8s context to user")
	if err := SwitchKubeContextForTest(t, "user-with-restricted-access"); err != nil {
		t.Errorf("failed to switch k8s context, %v", err)
		return
	}

	extraArgs := []string{"--set", fmt.Sprintf("mizu-resources-namespace=%s", namespace)}
	t.Run("basic tap", func (testingT *testing.T) {basicTapTest(testingT, false, extraArgs...)})
}
