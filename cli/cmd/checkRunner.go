package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/shared/semver"
	"net/http"
)

func runMizuCheck() {
	logger.Log.Infof("Mizu install checks\n===================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	kubernetesProvider, kubernetesVersion, checkPassed := checkKubernetesApi()

	var isInstallCommand bool
	if checkPassed {
		checkPassed, isInstallCommand = validateCommand(ctx, kubernetesProvider)
	}

	if checkPassed {
		checkPassed = checkKubernetesVersion(kubernetesVersion)
	}

	if checkPassed {
		checkPassed = checkAllResourcesExist(ctx, kubernetesProvider, isInstallCommand)
	}

	if checkPassed {
		checkPassed = checkServerConnection(kubernetesProvider, cancel)
	}

	if checkPassed {
		logger.Log.Infof("\nStatus check results are %v", fmt.Sprintf(uiUtils.Green, "√"))
	} else {
		logger.Log.Errorf("\nStatus check results are %v", fmt.Sprintf(uiUtils.Red, "✗"))
	}
}

func checkKubernetesApi() (*kubernetes.Provider, *semver.SemVersion, bool) {
	logger.Log.Infof("\nkubernetes-api\n--------------------")

	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		logger.Log.Errorf("%v can't initialize the client, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return nil, nil, false
	}
	logger.Log.Infof("%v can initialize the client", fmt.Sprintf(uiUtils.Green, "√"))

	if err := kubernetesProvider.ValidateNotProxy(); err != nil {
		logger.Log.Errorf("%v mizu doesn't support running through a proxy to k8s server, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return nil, nil, false
	}
	logger.Log.Infof("%v not running mizu through a proxy to k8s server", fmt.Sprintf(uiUtils.Green, "√"))

	kubernetesVersion, err := kubernetesProvider.GetKubernetesVersion()
	if err != nil {
		logger.Log.Errorf("%v can't query the Kubernetes API, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return nil, nil, false
	}
	logger.Log.Infof("%v can query the Kubernetes API", fmt.Sprintf(uiUtils.Green, "√"))

	return kubernetesProvider, kubernetesVersion, true
}

func validateCommand(ctx context.Context, kubernetesProvider *kubernetes.Provider) (bool, bool) {
	logger.Log.Infof("\nvalidate-mizu-command\n--------------------")

	if exist, err := kubernetesProvider.DoesDeploymentExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		logger.Log.Errorf("%v can't validate mizu command, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false, false
	} else if exist {
		logger.Log.Infof("%v mizu running with install command", fmt.Sprintf(uiUtils.Green, "√"))
		return true, true
	} else {
		logger.Log.Infof("%v mizu running with tap command", fmt.Sprintf(uiUtils.Green, "√"))
		return true, false
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

func checkServerConnection(kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) bool {
	logger.Log.Infof("\nmizu-connectivity\n--------------------")

	serverUrl := GetApiServerUrl()

	if response, err := http.Get(fmt.Sprintf("%s/", serverUrl)); err != nil || response.StatusCode != 200 {
		startProxyReportErrorIfAny(kubernetesProvider, cancel)
	}

	apiServerProvider := apiserver.NewProvider(serverUrl, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(); err != nil {
		logger.Log.Errorf("%v couldn't connect to API server, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	logger.Log.Infof("%v connected successfully to API server", fmt.Sprintf(uiUtils.Green, "√"))
	return true
}

func checkAllResourcesExist(ctx context.Context, kubernetesProvider *kubernetes.Provider, isInstallCommand bool) bool {
	logger.Log.Infof("\nmizu-existence\n--------------------")

	exist, err := kubernetesProvider.DoesNamespaceExist(ctx, config.Config.MizuResourcesNamespace)
	allResourcesExist := checkResourceExist(config.Config.MizuResourcesNamespace, "namespace", exist, err)

	exist, err = kubernetesProvider.DoesConfigMapExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ConfigMapName)
	allResourcesExist = checkResourceExist(kubernetes.ConfigMapName, "config map", exist, err)

	exist, err = kubernetesProvider.DoesServiceAccountExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName)
	allResourcesExist = checkResourceExist(kubernetes.ServiceAccountName, "service account", exist, err)

	if config.Config.IsNsRestrictedMode() {
		exist, err = kubernetesProvider.DoesRoleExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.RoleName)
		allResourcesExist = checkResourceExist(kubernetes.RoleName, "role", exist, err)

		exist, err = kubernetesProvider.DoesRoleBindingExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.RoleBindingName)
		allResourcesExist = checkResourceExist(kubernetes.RoleBindingName, "role binding", exist, err)
	} else {
		exist, err = kubernetesProvider.DoesClusterRoleExist(ctx, kubernetes.ClusterRoleName)
		allResourcesExist = checkResourceExist(kubernetes.ClusterRoleName, "cluster role", exist, err)

		exist, err = kubernetesProvider.DoesClusterRoleBindingExist(ctx, kubernetes.ClusterRoleBindingName)
		allResourcesExist = checkResourceExist(kubernetes.ClusterRoleBindingName, "cluster role binding", exist, err)
	}

	exist, err = kubernetesProvider.DoesServiceExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName)
	allResourcesExist = checkResourceExist(kubernetes.ApiServerPodName, "service", exist, err)

	if isInstallCommand {
		allResourcesExist = checkInstallResourcesExist(ctx, kubernetesProvider)
	} else {
		allResourcesExist = checkTapResourcesExist(ctx, kubernetesProvider)
	}

	return allResourcesExist
}

func checkInstallResourcesExist(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	exist, err := kubernetesProvider.DoesRoleExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.DaemonRoleName)
	installResourcesExist := checkResourceExist(kubernetes.DaemonRoleName, "role", exist, err)

	exist, err = kubernetesProvider.DoesRoleBindingExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.DaemonRoleBindingName)
	installResourcesExist = checkResourceExist(kubernetes.DaemonRoleBindingName, "role binding", exist, err)

	exist, err = kubernetesProvider.DoesPersistentVolumeClaimExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.PersistentVolumeClaimName)
	installResourcesExist = checkResourceExist(kubernetes.PersistentVolumeClaimName, "persistent volume claim", exist, err)

	exist, err = kubernetesProvider.DoesDeploymentExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName)
	installResourcesExist = checkResourceExist(kubernetes.ApiServerPodName, "deployment", exist, err)

	return installResourcesExist
}

func checkTapResourcesExist(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	exist, err := kubernetesProvider.DoesPodExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName)
	tapResourcesExist := checkResourceExist(kubernetes.ApiServerPodName, "pod", exist, err)

	return tapResourcesExist
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
