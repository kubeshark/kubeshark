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
			resource := "namespaces"
			group := ""
			permission, err := kubernetesProvider.CanI(ctx, namespace, resource, namespaceVerb, group)
			allPermissionsExist = checkResourcePermission(resource, namespaceVerb, group, permission, err) && allPermissionsExist
		}

		clusterRoleVerbs := []string{"get", "delete", "list", "create"}
		for _, clusterRoleVerb := range clusterRoleVerbs {
			resource := "clusterroles"
			group := "rbac.authorization.k8s.io"
			permission, err := kubernetesProvider.CanI(ctx, namespace, resource, clusterRoleVerb, group)
			allPermissionsExist = checkResourcePermission(resource, clusterRoleVerb, group, permission, err) && allPermissionsExist
		}

		clusterRoleBindingVerbs := []string{"get", "delete", "list", "create"}
		for _, clusterRoleBindingVerb := range clusterRoleBindingVerbs {
			resource := "clusterrolebindings"
			group := "rbac.authorization.k8s.io"
			permission, err := kubernetesProvider.CanI(ctx, namespace, resource, clusterRoleBindingVerb, group)
			allPermissionsExist = checkResourcePermission(resource, clusterRoleBindingVerb, group, permission, err) && allPermissionsExist
		}
	}

	configMapVerbs := []string{"get", "delete", "create"}
	for _, configMapVerb := range configMapVerbs {
		resource := "configmaps"
		group := ""
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, configMapVerb, group)
		allPermissionsExist = checkResourcePermission(resource, configMapVerb, group, permission, err) && allPermissionsExist
	}

	serviceAccountVerbs := []string{"get", "delete", "list", "create"}
	for _, serviceAccountVerb := range serviceAccountVerbs {
		resource := "serviceaccounts"
		group := ""
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, serviceAccountVerb, group)
		allPermissionsExist = checkResourcePermission(resource, serviceAccountVerb, group, permission, err) && allPermissionsExist
	}

	roleVerbs := []string{"get", "delete", "list", "create"}
	for _, roleVerb := range roleVerbs {
		resource := "roles"
		group := "rbac.authorization.k8s.io"
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, roleVerb, group)
		allPermissionsExist = checkResourcePermission(resource, roleVerb, group, permission, err) && allPermissionsExist
	}

	roleBindingVerbs := []string{"get", "delete", "list", "create"}
	for _, roleBindingVerb := range roleBindingVerbs {
		resource := "rolebindings"
		group := "rbac.authorization.k8s.io"
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, roleBindingVerb, group)
		allPermissionsExist = checkResourcePermission(resource, roleBindingVerb, group, permission, err) && allPermissionsExist
	}

	serviceVerbs := []string{"get", "delete", "create", "watch"}
	for _, serviceVerb := range serviceVerbs {
		resource := "services"
		group := ""
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, serviceVerb, group)
		allPermissionsExist = checkResourcePermission(resource, serviceVerb, group, permission, err) && allPermissionsExist
	}

	podVerbs := []string{"get", "delete", "list", "watch"}
	for _, podVerb := range podVerbs {
		resource := "pods"
		group := ""
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, podVerb, group)
		allPermissionsExist = checkResourcePermission(resource, podVerb, group, permission, err) && allPermissionsExist
	}

	daemonSetVerbs := []string{"delete", "create", "patch"}
	for _, daemonSetVerb := range daemonSetVerbs {
		resource := "daemonsets"
		group := "apps"
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, daemonSetVerb, group)
		allPermissionsExist = checkResourcePermission(resource, daemonSetVerb, group, permission, err) && allPermissionsExist
	}

	eventVerbs := []string{"list", "watch"}
	for _, eventVerb := range eventVerbs {
		resource := "events"
		group := "events.k8s.io"
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, eventVerb, group)
		allPermissionsExist = checkResourcePermission(resource, eventVerb, group, permission, err) && allPermissionsExist
	}

	endpointVerbs := []string{"watch"}
	for _, endpointVerb := range endpointVerbs {
		resource := "endpoints"
		group := ""
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, endpointVerb, group)
		allPermissionsExist = checkResourcePermission(resource, endpointVerb, group, permission, err) && allPermissionsExist
	}

	proxyVerbs := []string{"get"}
	for _, proxyVerb := range proxyVerbs {
		resource := "services/proxy"
		group := ""
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, proxyVerb, group)
		allPermissionsExist = checkResourcePermission(resource, proxyVerb, group, permission, err) && allPermissionsExist
	}

	portForwardVerbs := []string{"create"}
	for _, portForwardVerb := range portForwardVerbs {
		resource := "pods/portforward"
		group := ""
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, portForwardVerb, group)
		allPermissionsExist = checkResourcePermission(resource, portForwardVerb, group, permission, err) && allPermissionsExist
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
		resource := "persistentvolumeclaims"
		group := ""
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, persistentVolumeClaimVerb, group)
		installPermissionsExist = checkResourcePermission(resource, persistentVolumeClaimVerb, group, permission, err) && installPermissionsExist
	}

	deploymentVerbs := []string{"get", "delete", "create"}
	for _, deploymentVerb := range deploymentVerbs {
		resource := "deployments"
		group := "apps"
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, deploymentVerb, group)
		installPermissionsExist = checkResourcePermission(resource, deploymentVerb, group, permission, err) && installPermissionsExist
	}

	return installPermissionsExist
}

func checkTapResourcesPermission(ctx context.Context, kubernetesProvider *kubernetes.Provider, namespace string) bool {
	tapPermissionsExist := true

	podVerbs := []string{"create"}
	for _, podVerb := range podVerbs {
		resource := "pods"
		group := ""
		permission, err := kubernetesProvider.CanI(ctx, namespace, resource, podVerb, group)
		tapPermissionsExist = checkResourcePermission(resource, podVerb, group, permission, err) && tapPermissionsExist
	}

	return tapPermissionsExist
}

func checkResourcePermission(resource string, verb string, group string, permission bool, err error) bool {
	if err != nil {
		logger.Log.Errorf("%v error checking if permission for %v %v in group '%v' exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource, group, err)
		return false
	} else if !permission {
		logger.Log.Errorf("%v can't %v %v in group '%v'", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource, group)
		return false
	} else {
		logger.Log.Infof("%v can %v %v in group '%v'", fmt.Sprintf(uiUtils.Green, "√"), verb, resource, group)
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
