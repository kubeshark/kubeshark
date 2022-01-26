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
	"regexp"
)

func runMizuCheck() {
	logger.Log.Infof("Mizu install checks\n===================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	kubernetesProvider, kubernetesVersion, checkPassed := checkKubernetesApi()

	if checkPassed {
		checkPassed = checkKubernetesVersion(kubernetesVersion)
	}

	if config.Config.Check.PreInstall || config.Config.Check.PreTap {
		if checkPassed {
			checkPassed = checkAllResourcesPermission(ctx, kubernetesProvider, config.Config.Check.PreInstall, config.Config.Check.PreTap)
		}
	} else {
		var isInstallCommand bool
		if checkPassed {
			checkPassed, isInstallCommand = checkMizuMode(ctx, kubernetesProvider)
		}

		if checkPassed {
			checkPassed = checkAllResourcesExist(ctx, kubernetesProvider, isInstallCommand)
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

func checkMizuMode(ctx context.Context, kubernetesProvider *kubernetes.Provider) (bool, bool) {
	logger.Log.Infof("\nmizu-mode\n--------------------")

	if exist, err := kubernetesProvider.DoesDeploymentExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		logger.Log.Errorf("%v can't check mizu command, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
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

func checkServerConnection(kubernetesProvider *kubernetes.Provider) bool {
	logger.Log.Infof("\nmizu-connectivity\n--------------------")

	serverUrl := GetApiServerUrl()

	apiServerProvider := apiserver.NewProviderWithoutRetries(serverUrl, apiserver.DefaultTimeout)
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

func checkAllResourcesPermission(ctx context.Context, kubernetesProvider *kubernetes.Provider, checkInstall bool, checkTap bool) bool {
	logger.Log.Infof("\nmizu-resource-permission\n--------------------")

	allPermissionsExist := true

	var namespace string
	if config.Config.IsNsRestrictedMode() {
		namespace = config.Config.MizuResourcesNamespace
	} else {
		namespace = ""
	}

	if !config.Config.IsNsRestrictedMode() {
		namespaceVerbs := []string{"get", "delete", "list", "create", "watch"}
		for _, namespaceVerb := range namespaceVerbs {
			permission, err := kubernetesProvider.CanI(ctx, namespace, "namespaces", namespaceVerb)
			allPermissionsExist = checkResourcePermission("namespaces", namespaceVerb, permission, err) && allPermissionsExist
		}

		clusterRoleVerbs := []string{"get", "delete", "list", "create"}
		for _, clusterRoleVerb := range clusterRoleVerbs {
			permission, err := kubernetesProvider.CanI(ctx, namespace, "clusterroles", clusterRoleVerb)
			allPermissionsExist = checkResourcePermission("clusterroles", clusterRoleVerb, permission, err) && allPermissionsExist
		}

		clusterRoleBindingVerbs := []string{"get", "delete", "list", "create"}
		for _, clusterRoleBindingVerb := range clusterRoleBindingVerbs {
			permission, err := kubernetesProvider.CanI(ctx, namespace, "clusterrolebindings", clusterRoleBindingVerb)
			allPermissionsExist = checkResourcePermission("clusterrolebindings", clusterRoleBindingVerb, permission, err) && allPermissionsExist
		}
	}

	configMapVerbs := []string{"get", "delete", "create"}
	for _, configMapVerb := range configMapVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "configmaps", configMapVerb)
		allPermissionsExist = checkResourcePermission("configmaps", configMapVerb, permission, err) && allPermissionsExist
	}

	serviceAccountVerbs := []string{"get", "delete", "list", "create"}
	for _, serviceAccountVerb := range serviceAccountVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "serviceaccounts", serviceAccountVerb)
		allPermissionsExist = checkResourcePermission("serviceaccounts", serviceAccountVerb, permission, err) && allPermissionsExist
	}

	roleVerbs := []string{"get", "delete", "list", "create"}
	for _, roleVerb := range roleVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "roles", roleVerb)
		allPermissionsExist = checkResourcePermission("roles", roleVerb, permission, err) && allPermissionsExist
	}

	roleBindingVerbs := []string{"get", "delete", "list", "create"}
	for _, roleBindingVerb := range roleBindingVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "rolebindings", roleBindingVerb)
		allPermissionsExist = checkResourcePermission("rolebindings", roleBindingVerb, permission, err) && allPermissionsExist
	}

	serviceVerbs := []string{"get", "delete", "create", "watch"}
	for _, serviceVerb := range serviceVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "services", serviceVerb)
		allPermissionsExist = checkResourcePermission("services", serviceVerb, permission, err) && allPermissionsExist
	}

	podVerbs := []string{"get", "delete", "list", "watch"}
	for _, podVerb := range podVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "pods", podVerb)
		allPermissionsExist = checkResourcePermission("pods", podVerb, permission, err) && allPermissionsExist
	}

	daemonSetVerbs := []string{"delete", "create", "patch"}
	for _, daemonSetVerb := range daemonSetVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "daemonsets", daemonSetVerb)
		allPermissionsExist = checkResourcePermission("daemonsets", daemonSetVerb, permission, err) && allPermissionsExist
	}

	eventVerbs := []string{"list", "watch"}
	for _, eventVerb := range eventVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "events", eventVerb)
		allPermissionsExist = checkResourcePermission("events", eventVerb, permission, err) && allPermissionsExist
	}

	endpointVerbs := []string{"watch"}
	for _, endpointVerb := range endpointVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "endpoints", endpointVerb)
		allPermissionsExist = checkResourcePermission("endpoints", endpointVerb, permission, err) && allPermissionsExist
	}

	if checkInstall {
		allPermissionsExist = checkInstallResourcesPermission(ctx, kubernetesProvider, namespace) && allPermissionsExist
	}

	if checkTap {
		allPermissionsExist = checkTapResourcesPermission(ctx, kubernetesProvider, namespace) && allPermissionsExist
	}

	return allPermissionsExist
}

func checkInstallResourcesPermission(ctx context.Context, kubernetesProvider *kubernetes.Provider, namespace string) bool {
	installPermissionsExist := true

	persistentVolumeClaimVerbs := []string{"get", "delete", "create"}
	for _, persistentVolumeClaimVerb := range persistentVolumeClaimVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "persistentvolumeclaims", persistentVolumeClaimVerb)
		installPermissionsExist = checkResourcePermission("persistentvolumeclaims", persistentVolumeClaimVerb, permission, err) && installPermissionsExist
	}

	deploymentVerbs := []string{"get", "delete", "create"}
	for _, deploymentVerb := range deploymentVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "deployments", deploymentVerb)
		installPermissionsExist = checkResourcePermission("deployments", deploymentVerb, permission, err) && installPermissionsExist
	}

	return installPermissionsExist
}

func checkTapResourcesPermission(ctx context.Context, kubernetesProvider *kubernetes.Provider, namespace string) bool {
	tapPermissionsExist := true

	podVerbs := []string{"create"}
	for _, podVerb := range podVerbs {
		permission, err := kubernetesProvider.CanI(ctx, namespace, "pods", podVerb)
		tapPermissionsExist = checkResourcePermission("pods", podVerb, permission, err) && tapPermissionsExist
	}

	return tapPermissionsExist
}

func checkResourcePermission(resource string, verb string, permission bool, err error) bool {
	if err != nil {
		logger.Log.Errorf("%v error checking if permission for %v %v exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource, err)
		return false
	} else if !permission {
		logger.Log.Errorf("%v can't %v %v", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource)
		return false
	} else {
		logger.Log.Infof("%v can %v %v", fmt.Sprintf(uiUtils.Green, "√"), verb, resource)
	}

	return true
}

func checkAllResourcesExist(ctx context.Context, kubernetesProvider *kubernetes.Provider, isInstallCommand bool) bool {
	logger.Log.Infof("\nmizu-existence\n--------------------")

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
