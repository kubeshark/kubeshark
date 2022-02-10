package cmd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/shared/semver"
)

func runMizuCheck() {
	logger.Log.Infof("Mizu install checks\n===================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	kubernetesProvider, kubernetesVersion, checkPassed := checkKubernetesApi()

	if checkPassed {
		checkPassed = checkKubernetesVersion(kubernetesVersion)
	}

	var isInstallCommand bool
	if checkPassed {
		checkPassed, isInstallCommand = checkMizuMode(ctx, kubernetesProvider)
	}

	if checkPassed {
		checkPassed = checkK8sResources(ctx, kubernetesProvider, isInstallCommand)
	}

	if checkPassed {
		checkPassed = checkServerConnection(kubernetesProvider)
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

func checkMizuMode(ctx context.Context, kubernetesProvider *kubernetes.Provider) (bool, bool) {
	logger.Log.Infof("\nmode\n--------------------")

	if exist, err := kubernetesProvider.DoesDeploymentExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		logger.Log.Errorf("%v can't check mizu command, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false, false
	} else if exist {
		logger.Log.Infof("%v mizu running with install command", fmt.Sprintf(uiUtils.Green, "√"))
		return true, true
	} else if exist, err = kubernetesProvider.DoesPodExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		logger.Log.Errorf("%v can't check mizu command, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false, false
	} else if exist {
		logger.Log.Infof("%v mizu running with tap command", fmt.Sprintf(uiUtils.Green, "√"))
		return true, false
	} else {
		logger.Log.Infof("%v mizu is not running", fmt.Sprintf(uiUtils.Red, "✗"))
		return false, false
	}
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

	serverUrl := GetApiServerUrl()

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

func checkK8sResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, isInstallCommand bool) bool {
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

	if isInstallCommand {
		allResourcesExist = checkInstallResourcesExist(ctx, kubernetesProvider) && allResourcesExist
	} else {
		allResourcesExist = checkTapResourcesExist(ctx, kubernetesProvider) && allResourcesExist
	}

	return allResourcesExist
}

func checkInstallResourcesExist(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	exist, err := kubernetesProvider.DoesRoleExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.DaemonRoleName)
	installResourcesExist := checkResourceExist(kubernetes.DaemonRoleName, "role", exist, err)

	exist, err = kubernetesProvider.DoesRoleBindingExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.DaemonRoleBindingName)
	installResourcesExist = checkResourceExist(kubernetes.DaemonRoleBindingName, "role binding", exist, err) && installResourcesExist

	exist, err = kubernetesProvider.DoesPersistentVolumeClaimExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.PersistentVolumeClaimName)
	installResourcesExist = checkResourceExist(kubernetes.PersistentVolumeClaimName, "persistent volume claim", exist, err) && installResourcesExist

	exist, err = kubernetesProvider.DoesDeploymentExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName)
	installResourcesExist = checkResourceExist(kubernetes.ApiServerPodName, "deployment", exist, err) && installResourcesExist

	return installResourcesExist
}

func checkTapResourcesExist(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
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
		if tappers != notRunningTappers {
			logger.Log.Errorf("%v '%v' %v/%v pods are not running", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.TapperPodName, notRunningTappers, tappers)
			return false
		}
		logger.Log.Infof("%v %v '%v' pods running", fmt.Sprintf(uiUtils.Green, "√"), tappers, kubernetes.TapperPodName)
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
	} else {
		logger.Log.Infof("%v '%v' %v exists", fmt.Sprintf(uiUtils.Green, "√"), resourceName, resourceType)
	}

	return true
}
