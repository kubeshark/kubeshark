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
	LongRetriesCount      = 100
	ShortRetriesCount     = 10
	DefaultApiServerPort  = shared.DefaultApiServerPort
	DefaultNamespaceName  = "mizu-tests"
	DefaultServiceName    = "httpbin"
	DefaultEntriesCount   = 50
	WaitAfterTapPodsReady = 3 * time.Second
	AllNamespaces         = ""
)

type PodDescriptor struct {
	Name      string
	Namespace string
}

func GetCliPath() (string, error) {
	dir, filePathErr := os.Getwd()
	if filePathErr != nil {
		return "", filePathErr
	}

	cliPath := path.Join(dir, "../cli/bin/mizu_ci")
	return cliPath, nil
}

func GetMizuFolderPath() (string, error) {
	home, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return "", homeDirErr
	}

	return path.Join(home, ".mizu"), nil
}

func GetConfigPath() (string, error) {
	mizuFolderPath, mizuPathError := GetMizuFolderPath()
	if mizuPathError != nil {
		return "", mizuPathError
	}

	return path.Join(mizuFolderPath, "config.yaml"), nil
}

func GetProxyUrl(namespace string, service string) string {
	return fmt.Sprintf("http://localhost:8080/api/v1/namespaces/%v/services/%v/proxy", namespace, service)
}

func GetApiServerUrl(port uint16) string {
	return fmt.Sprintf("http://localhost:%v", port)
}

func NewKubernetesProvider() (*KubernetesProvider, error) {
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
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(restClientConfig)
	if err != nil {
		return nil, err
	}

	return &KubernetesProvider{clientSet}, nil
}

type KubernetesProvider struct {
	clientSet *kubernetes.Clientset
}

func (kp *KubernetesProvider) GetServiceExternalIp(ctx context.Context, namespace string, service string) (string, error) {
	serviceObj, err := kp.clientSet.CoreV1().Services(namespace).Get(ctx, service, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	externalIp := serviceObj.Status.LoadBalancer.Ingress[0].IP
	return externalIp, nil
}

func SwitchKubeContextForTest(t *testing.T, newContextName string) error {
	prevKubeContextName, err := GetKubeCurrentContextName()
	if err != nil {
		return err
	}

	if err := SetKubeCurrentContext(newContextName); err != nil {
		return err
	}

	t.Cleanup(func() {
		if err := SetKubeCurrentContext(prevKubeContextName); err != nil {
			t.Errorf("failed to set Kubernetes context to %s, err: %v", prevKubeContextName, err)
			t.Errorf("cleanup failed, subsequent tests may be affected")
		}
	})

	return nil
}

func GetKubeCurrentContextName() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v, %s", err, string(output))
	}

	return string(bytes.TrimSpace(output)), nil
}

func SetKubeCurrentContext(contextName string) error {
	cmd := exec.Command("kubectl", "config", "use-context", contextName)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v, %s", err, string(output))
	}

	return nil
}

func ApplyKubeFilesForTest(t *testing.T, kubeContext string, namespace string, filename ...string) error {
	for i := range filename {
		fname := filename[i]
		if err := ApplyKubeFile(kubeContext, namespace, fname); err != nil {
			return err
		}

		t.Cleanup(func() {
			if err := DeleteKubeFile(kubeContext, namespace, fname); err != nil {
				t.Errorf(
					"failed to delete Kubernetes resources in namespace %s from filename %s, err: %v",
					namespace,
					fname,
					err,
				)
			}
		})
	}

	return nil
}

func ApplyKubeFile(kubeContext string, namespace string, filename string) (error) {
	cmdArgs := []string{
		"apply",
		"--context", kubeContext,
		"-f", filename,
	}
	if namespace != AllNamespaces {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}
	cmd := exec.Command("kubectl", cmdArgs...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v, %s", err, string(output))
	}

	return nil
}

func DeleteKubeFile(kubeContext string, namespace string, filename string) error {
	cmdArgs := []string{
		"delete",
		"--context", kubeContext,
		"-f", filename,
	}
	if namespace != AllNamespaces {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}
	cmd := exec.Command("kubectl", cmdArgs...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v, %s", err, string(output))
	}

	return nil
}

func getDefaultCommandArgs() []string {
	agentImageValue := os.Getenv("MIZU_CI_IMAGE")
	setFlag := "--set"
	telemetry := "telemetry=false"
	agentImage := fmt.Sprintf("agent-image=%s", agentImageValue)
	imagePullPolicy := "image-pull-policy=IfNotPresent"
	headless := "headless=true"

	return []string{setFlag, telemetry, setFlag, agentImage, setFlag, imagePullPolicy, setFlag, headless}
}

func GetDefaultTapCommandArgs() []string {
	tapCommand := "tap"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{tapCommand}, defaultCmdArgs...)
}

func GetDefaultTapCommandArgsWithRegex(regex string) []string {
	tapCommand := "tap"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{tapCommand, regex}, defaultCmdArgs...)
}

func GetDefaultLogsCommandArgs() []string {
	logsCommand := "logs"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{logsCommand}, defaultCmdArgs...)
}

func GetDefaultTapNamespace() []string {
	return []string{"-n", "mizu-tests"}
}

func GetDefaultConfigCommandArgs() []string {
	configCommand := "config"
	defaultCmdArgs := getDefaultCommandArgs()

	return append([]string{configCommand}, defaultCmdArgs...)
}

func RunCypressTests(t *testing.T, cypressRunCmd string) {
	cypressCmd := exec.Command("bash", "-c", cypressRunCmd)
	t.Logf("running command: %v", cypressCmd.String())
	out, err := cypressCmd.CombinedOutput()
	if err != nil {
		t.Errorf("error running cypress, error: %v, output: %v", err, string(out))
		return
	}

	t.Logf("%s", out)
}

func RetriesExecute(retriesCount int, executeFunc func() error) error {
	var lastError interface{}

	for i := 0; i < retriesCount; i++ {
		if err := TryExecuteFunc(executeFunc); err != nil {
			lastError = err

			time.Sleep(1 * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("reached max retries count, retries count: %v, last err: %v", retriesCount, lastError)
}

func TryExecuteFunc(executeFunc func() error) (err interface{}) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = panicErr
		}
	}()

	return executeFunc()
}

func WaitTapPodsReady(apiServerUrl string) error {
	resolvingUrl := fmt.Sprintf("%v/status/connectedTappersCount", apiServerUrl)
	tapPodsReadyFunc := func() error {
		requestResult, requestErr := ExecuteHttpGetRequest(resolvingUrl)
		if requestErr != nil {
			return requestErr
		}

		connectedTappersCount := requestResult.(float64)
		if connectedTappersCount == 0 {
			return fmt.Errorf("no connected tappers running")
		}
		time.Sleep(WaitAfterTapPodsReady)
		return nil
	}

	return RetriesExecute(LongRetriesCount, tapPodsReadyFunc)
}

func JsonBytesToInterface(jsonBytes []byte) (interface{}, error) {
	var result interface{}
	if parseErr := json.Unmarshal(jsonBytes, &result); parseErr != nil {
		return nil, parseErr
	}

	return result, nil
}

func ExecuteHttpRequest(response *http.Response, requestErr error) (interface{}, error) {
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

	return JsonBytesToInterface(data)
}

func ExecuteHttpGetRequestWithHeaders(url string, headers map[string]string) (interface{}, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for headerKey, headerValue := range headers {
		request.Header.Add(headerKey, headerValue)
	}

	client := &http.Client{}
	response, requestErr := client.Do(request)
	return ExecuteHttpRequest(response, requestErr)
}

func ExecuteHttpGetRequest(url string) (interface{}, error) {
	response, requestErr := http.Get(url)
	return ExecuteHttpRequest(response, requestErr)
}

func ExecuteHttpPostRequestWithHeaders(url string, headers map[string]string, body interface{}) (interface{}, error) {
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
	return ExecuteHttpRequest(response, requestErr)
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

func GetLogsPath() (string, error) {
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
