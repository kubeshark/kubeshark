package acceptanceTests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/up9inc/mizu/shared"
)

const (
	longRetriesCount      = 100
	shortRetriesCount     = 10
	defaultApiServerPort  = shared.DefaultApiServerPort
	defaultNamespaceName  = "mizu-tests"
	defaultServiceName    = "httpbin"
	defaultEntriesCount   = 50
	waitAfterTapPodsReady = 3 * time.Second
)

type PodDescriptor struct {
	Name      string
	Namespace string
}

func getCliPath() (string, error) {
	dir, filePathErr := os.Getwd()
	if filePathErr != nil {
		return "", filePathErr
	}

	cliPath := path.Join(dir, "../cli/bin/mizu_ci")
	return cliPath, nil
}

func getMizuFolderPath() (string, error) {
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return "", homeDirErr
	}

	return path.Join(home, ".mizu"), nil
}

func getConfigPath() (string, error) {
	mizuFolderPath, mizuPathError := getMizuFolderPath()
	if mizuPathError != nil {
		return "", mizuPathError
	}

	return path.Join(mizuFolderPath, "config.yaml"), nil
}

func getProxyUrl(namespace string, service string) string {
	return fmt.Sprintf("http://localhost:8080/api/v1/namespaces/%v/services/%v/proxy", namespace, service)
}

func getApiServerUrl(port uint16) string {
	return fmt.Sprintf("http://localhost:%v", port)
}

func getServiceExternalIp(ctx context.Context, namespace string, service string) (string, error) {
	home := homedir.HomeDir()
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: filepath.Join(home, ".kube", "config")}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		},
	)

	restClientConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return "", err
	}

	clientSet, err := kubernetes.NewForConfig(restClientConfig)
	if err != nil {
		return "", err
	}

	serviceObj, err := clientSet.CoreV1().Services(namespace).Get(ctx, service, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	externalIp := serviceObj.Status.LoadBalancer.Ingress[0].IP
	return externalIp, nil
}

func getDefaultCommandArgs() []string {
	setFlag := "--set"
	telemetry := "telemetry=false"
	agentImage := "agent-image=gcr.io/up9-docker-hub/mizu/ci:0.0"
	imagePullPolicy := "image-pull-policy=IfNotPresent"
	headless := "headless=true"

	return []string{setFlag, telemetry, setFlag, agentImage, setFlag, imagePullPolicy, setFlag, headless}
}

func getDefaultTapCommandArgs() []string {
	tapCommand := "tap"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{tapCommand}, defaultCmdArgs...)
}

func getDefaultTapCommandArgsWithRegex(regex string) []string {
	tapCommand := "tap"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{tapCommand, regex}, defaultCmdArgs...)
}

func getDefaultLogsCommandArgs() []string {
	logsCommand := "logs"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{logsCommand}, defaultCmdArgs...)
}

func getDefaultTapNamespace() []string {
	return []string{"-n", "mizu-tests"}
}

func getDefaultConfigCommandArgs() []string {
	configCommand := "config"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{configCommand}, defaultCmdArgs...)
}

func runCypressTests(t *testing.T, cypressRunCmd string) {
	cypressCmd := exec.Command("bash", "-c", cypressRunCmd)
	t.Logf("running command: %v", cypressCmd.String())
	out, err := cypressCmd.Output()
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}

	t.Logf("%s", out)
}

func retriesExecute(retriesCount int, executeFunc func() error) error {
	var lastError interface{}

	for i := 0; i < retriesCount; i++ {
		if err := tryExecuteFunc(executeFunc); err != nil {
			lastError = err

			time.Sleep(1 * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("reached max retries count, retries count: %v, last err: %v", retriesCount, lastError)
}

func tryExecuteFunc(executeFunc func() error) (err interface{}) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = panicErr
		}
	}()

	return executeFunc()
}

func waitTapPodsReady(apiServerUrl string) error {
	resolvingUrl := fmt.Sprintf("%v/status/connectedTappersCount", apiServerUrl)
	tapPodsReadyFunc := func() error {
		requestResult, requestErr := executeHttpGetRequest(resolvingUrl)
		if requestErr != nil {
			return requestErr
		}

		connectedTappersCount := requestResult.(float64)
		if connectedTappersCount == 0 {
			return fmt.Errorf("no connected tappers running")
		}
		time.Sleep(waitAfterTapPodsReady)
		return nil
	}

	return retriesExecute(longRetriesCount, tapPodsReadyFunc)
}

func jsonBytesToInterface(jsonBytes []byte) (interface{}, error) {
	var result interface{}
	if parseErr := json.Unmarshal(jsonBytes, &result); parseErr != nil {
		return nil, parseErr
	}

	return result, nil
}

func executeHttpRequest(response *http.Response, requestErr error) (interface{}, error) {
	if requestErr != nil {
		return nil, requestErr
	} else if response.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code %v", response.StatusCode)
	}

	defer func() { response.Body.Close() }()

	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	return jsonBytesToInterface(data)
}

func executeHttpGetRequestWithHeaders(url string, headers map[string]string) (interface{}, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for headerKey, headerValue := range headers {
		request.Header.Add(headerKey, headerValue)
	}

	client := &http.Client{}
	response, requestErr := client.Do(request)
	return executeHttpRequest(response, requestErr)
}

func executeHttpGetRequest(url string) (interface{}, error) {
	response, requestErr := http.Get(url)
	return executeHttpRequest(response, requestErr)
}

func executeHttpPostRequestWithHeaders(url string, headers map[string]string, body interface{}) (interface{}, error) {
	requestBody, jsonErr := json.Marshal(body)
	if jsonErr != nil {
		return nil, jsonErr
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	for headerKey, headerValue := range headers {
		request.Header.Add(headerKey, headerValue)
	}

	client := &http.Client{}
	response, requestErr := client.Do(request)
	return executeHttpRequest(response, requestErr)
}

func cleanupCommand(cmd *exec.Cmd) error {
	if err := cmd.Process.Signal(syscall.SIGQUIT); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func getLogsPath() (string, error) {
	dir, filePathErr := os.Getwd()
	if filePathErr != nil {
		return "", filePathErr
	}

	logsPath := path.Join(dir, "mizu_logs.zip")
	return logsPath, nil
}

func Contains(slice []string, containsValue string) bool {
	for _, sliceValue := range slice {
		if sliceValue == containsValue {
			return true
		}
	}

	return false
}

func ContainsPartOfValue(slice []string, containsValue string) bool {
	for _, sliceValue := range slice {
		if strings.Contains(sliceValue, containsValue) {
			return true
		}
	}

	return false
}
