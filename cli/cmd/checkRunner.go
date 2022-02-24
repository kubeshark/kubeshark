package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/shared"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"regexp"

	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/shared/semver"
)

func runMizuCheck() {
	logger.Log.Infof("Mizu checks\n===================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	kubernetesProvider, kubernetesVersion, checkPassed := checkKubernetesApi()

	if checkPassed {
		checkPassed = checkKubernetesVersion(kubernetesVersion)
	}

	if config.Config.Check.PreTap {
		if checkPassed {
			checkPassed = checkK8sTapPermissions(ctx, kubernetesProvider)
		}
	} else {
		if checkPassed {
			checkPassed = checkK8sResources(ctx, kubernetesProvider)
		}

		if checkPassed {
			checkPassed = checkServerConnection(kubernetesProvider)
		}
	}

	if checkPassed {
		logger.Log.Infof("\nStatus check results are %v", fmt.Sprintf(uiUtils.Green, "√"))
	} else {
		logger.Log.Errorf("\nStatus check results are %v", fmt.Sprintf(uiUtils.Red, "✗"))
	}
}

func checkKubernetesApi() (*kubernetes.Provider, *semver.SemVersion, bool) {
	logger.Log.Infof("\nkubernetes-api\n--------------------")

	kubernetesProvider, err := kubernetes.NewProvider(config.Config.KubeConfigPath())
	if err != nil {
		logger.Log.Errorf("%v can't initialize the client, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return nil, nil, false
	}
	logger.Log.Infof("%v can initialize the client", fmt.Sprintf(uiUtils.Green, "√"))

	kubernetesVersion, err := kubernetesProvider.GetKubernetesVersion()
	if err != nil {
		logger.Log.Errorf("%v can't query the Kubernetes API, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return nil, nil, false
	}
	logger.Log.Infof("%v can query the Kubernetes API", fmt.Sprintf(uiUtils.Green, "√"))

	return kubernetesProvider, kubernetesVersion, true
}

func checkKubernetesVersion(kubernetesVersion *semver.SemVersion) bool {
	logger.Log.Infof("\nkubernetes-version\n--------------------")

	if err := kubernetes.ValidateKubernetesVersion(kubernetesVersion); err != nil {
		logger.Log.Errorf("%v not running the minimum Kubernetes API version, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	logger.Log.Infof("%v is running the minimum Kubernetes API version", fmt.Sprintf(uiUtils.Green, "√"))
	return true
}

func checkServerConnection(kubernetesProvider *kubernetes.Provider) bool {
	logger.Log.Infof("\nAPI-server-connectivity\n--------------------")

	serverUrl := GetApiServerUrl(config.Config.Tap.GuiPort)

	apiServerProvider := apiserver.NewProvider(serverUrl, 1, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(); err == nil {
		logger.Log.Infof("%v found Mizu server tunnel available and connected successfully to API server", fmt.Sprintf(uiUtils.Green, "√"))
		return true
	}

	connectedToApiServer := false

	if err := checkProxy(serverUrl, kubernetesProvider); err != nil {
		logger.Log.Errorf("%v couldn't connect to API server using proxy, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
	} else {
		connectedToApiServer = true
		logger.Log.Infof("%v connected successfully to API server using proxy", fmt.Sprintf(uiUtils.Green, "√"))
	}

	if err := checkPortForward(serverUrl, kubernetesProvider); err != nil {
		logger.Log.Errorf("%v couldn't connect to API server using port-forward, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
	} else {
		connectedToApiServer = true
		logger.Log.Infof("%v connected successfully to API server using port-forward", fmt.Sprintf(uiUtils.Green, "√"))
	}

	return connectedToApiServer
}

func checkProxy(serverUrl string, kubernetesProvider *kubernetes.Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.ProxyHost, config.Config.Tap.GuiPort, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName, cancel)
	if err != nil {
		return err
	}

	apiServerProvider := apiserver.NewProvider(serverUrl, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(); err != nil {
		return err
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Log.Debugf("Error occurred while stopping proxy, err: %v", err)
	}

	return nil
}

func checkPortForward(serverUrl string, kubernetesProvider *kubernetes.Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podRegex, _ := regexp.Compile(kubernetes.ApiServerPodName)
	forwarder, err := kubernetes.NewPortForward(kubernetesProvider, config.Config.MizuResourcesNamespace, podRegex, config.Config.Tap.GuiPort, ctx, cancel)
	if err != nil {
		return err
	}

	apiServerProvider := apiserver.NewProvider(serverUrl, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(); err != nil {
		return err
	}

	forwarder.Close()

	return nil
}

func checkK8sResources(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	logger.Log.Infof("\nk8s-components\n--------------------")

	exist, err := kubernetesProvider.DoesNamespaceExist(ctx, config.Config.MizuResourcesNamespace)
	allResourcesExist := checkResourceExist(config.Config.MizuResourcesNamespace, "namespace", exist, err)

	exist, err = kubernetesProvider.DoesConfigMapExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ConfigMapName)
	allResourcesExist = checkResourceExist(kubernetes.ConfigMapName, "config map", exist, err) && allResourcesExist

	exist, err = kubernetesProvider.DoesServiceAccountExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName)
	allResourcesExist = checkResourceExist(kubernetes.ServiceAccountName, "service account", exist, err) && allResourcesExist

	if config.Config.IsNsRestrictedMode() {
		exist, err = kubernetesProvider.DoesRoleExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.RoleName)
		allResourcesExist = checkResourceExist(kubernetes.RoleName, "role", exist, err) && allResourcesExist

		exist, err = kubernetesProvider.DoesRoleBindingExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.RoleBindingName)
		allResourcesExist = checkResourceExist(kubernetes.RoleBindingName, "role binding", exist, err) && allResourcesExist
	} else {
		exist, err = kubernetesProvider.DoesClusterRoleExist(ctx, kubernetes.ClusterRoleName)
		allResourcesExist = checkResourceExist(kubernetes.ClusterRoleName, "cluster role", exist, err) && allResourcesExist

		exist, err = kubernetesProvider.DoesClusterRoleBindingExist(ctx, kubernetes.ClusterRoleBindingName)
		allResourcesExist = checkResourceExist(kubernetes.ClusterRoleBindingName, "cluster role binding", exist, err) && allResourcesExist
	}

	exist, err = kubernetesProvider.DoesServiceExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName)
	allResourcesExist = checkResourceExist(kubernetes.ApiServerPodName, "service", exist, err) && allResourcesExist

	allResourcesExist = checkPodResourcesExist(ctx, kubernetesProvider) && allResourcesExist

	return allResourcesExist
}

func checkPodResourcesExist(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	exist, err := kubernetesProvider.DoesPodExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName)
	tapResourcesExist := checkResourceExist(kubernetes.ApiServerPodName, "pod", exist, err)

	if !tapResourcesExist {
		return false
	}

	if pod, err := kubernetesProvider.GetPod(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' pod exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ApiServerPodName, err)
		return false
	} else if kubernetes.IsPodRunning(pod) {
		logger.Log.Infof("%v '%v' pod running", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.ApiServerPodName)
	} else {
		logger.Log.Errorf("%v '%v' pod not running", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ApiServerPodName)
		return false
	}

	tapperRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", kubernetes.TapperPodName))
	if pods, err := kubernetesProvider.ListAllPodsMatchingRegex(ctx, tapperRegex, []string{config.Config.MizuResourcesNamespace}); err != nil {
		logger.Log.Errorf("%v error listing '%v' pods, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.TapperPodName, err)
		return false
	} else {
		tappers := 0
		notRunningTappers := 0

		for _, pod := range pods {
			tappers += 1
			if !kubernetes.IsPodRunning(&pod) {
				notRunningTappers += 1
			}
		}

		if notRunningTappers > 0 {
			logger.Log.Errorf("%v '%v' %v/%v pods are not running", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.TapperPodName, notRunningTappers, tappers)
			return false
		}

		logger.Log.Infof("%v '%v' %v pods running", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.TapperPodName, tappers)
		return true
	}
}

func checkResourceExist(resourceName string, resourceType string, exist bool, err error) bool {
	if err != nil {
		logger.Log.Errorf("%v error checking if '%v' %v exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), resourceName, resourceType, err)
		return false
	} else if !exist {
		logger.Log.Errorf("%v '%v' %v doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), resourceName, resourceType)
		return false
	}

	logger.Log.Infof("%v '%v' %v exists", fmt.Sprintf(uiUtils.Green, "√"), resourceName, resourceType)
	return true
}

func checkK8sTapPermissions(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	logger.Log.Infof("\nkubernetes-permissions\n--------------------")

	var filePath string
	if config.Config.IsNsRestrictedMode() {
		filePath = "./examples/roles/permissions-ns-tap.yaml"
	} else {
		filePath = "./examples/roles/permissions-all-namespaces-tap.yaml"
	}

	data, err := shared.ReadFromFile(filePath)
	if err != nil {
		logger.Log.Errorf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	obj, err := getDecodedObject(data)
	if err != nil {
		logger.Log.Errorf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	var rules []rbac.PolicyRule
	if config.Config.IsNsRestrictedMode() {
		rules = obj.(*rbac.Role).Rules
	} else {
		rules = obj.(*rbac.ClusterRole).Rules
	}

	return checkPermissions(ctx, kubernetesProvider, rules)
}

func getDecodedObject(data []byte) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func checkPermissions(ctx context.Context, kubernetesProvider *kubernetes.Provider, rules []rbac.PolicyRule) bool {
	permissionsExist := true

	for _, rule := range rules {
		for _, group := range rule.APIGroups {
			for _, resource := range rule.Resources {
				for _, verb := range rule.Verbs {
					exist, err := kubernetesProvider.CanI(ctx, config.Config.MizuResourcesNamespace, resource, verb, group)
					permissionsExist = checkPermissionExist(group, resource, verb, exist, err) && permissionsExist
				}
			}
		}
	}

	return permissionsExist
}

func checkPermissionExist(group string, resource string, verb string, exist bool, err error) bool {
	if err != nil {
		logger.Log.Errorf("%v error checking permission for %v %v in group '%v', err: %v", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource, group, err)
		return false
	} else if !exist {
		logger.Log.Errorf("%v can't %v %v in group '%v'", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource, group)
		return false
	}

	logger.Log.Infof("%v can %v %v in group '%v'", fmt.Sprintf(uiUtils.Green, "√"), verb, resource, group)
	return true
}
