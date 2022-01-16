package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	"net/http"
)

func runMizuCheck() {
	logger.Log.Infof("Mizu install checks\n===================")

	kubernetesProvider, checkPassed := checkKubernetesApi()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	if checkPassed {
		checkPassed = checkAllResourcesExists(ctx, kubernetesProvider)
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

func checkKubernetesApi() (*kubernetes.Provider, bool) {
	logger.Log.Infof("\nkubernetes-api\n--------------------")

	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		logger.Log.Errorf("%v can't initialize the client, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return nil, false
	}

	logger.Log.Infof("%v can initialize the client", fmt.Sprintf(uiUtils.Green, "√"))
	return kubernetesProvider, true
}

func checkServerConnection(kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) bool {
	logger.Log.Infof("\nmizu-connectivity\n--------------------")

	serverUrl := config.Config.Check.ServerUrl

	if serverUrl == "" {
		serverUrl = GetApiServerUrl()

		if response, err := http.Get(fmt.Sprintf("%s/", serverUrl)); err != nil || response.StatusCode != 200 {
			go startProxyReportErrorIfAny(kubernetesProvider, cancel)
		}
	}

	apiServerProvider := apiserver.NewProvider(serverUrl, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(); err != nil {
		logger.Log.Errorf("%v couldn't connect to API server, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	logger.Log.Infof("%v connected successfully to API server", fmt.Sprintf(uiUtils.Green, "√"))
	return true
}

func checkAllResourcesExists(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	logger.Log.Infof("\nmizu-existence\n--------------------")

	allResourcesExists := true

	if doesResourceExist, err := kubernetesProvider.DoesNamespaceExist(ctx, config.Config.MizuResourcesNamespace); err != nil {
		logger.Log.Errorf("%v error checking if '%v' namespace exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), config.Config.MizuResourcesNamespace, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' namespace doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), config.Config.MizuResourcesNamespace)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' namespace exists", fmt.Sprintf(uiUtils.Green, "√"), config.Config.MizuResourcesNamespace)
	}

	if doesResourceExist, err := kubernetesProvider.DoesConfigMapExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ConfigMapName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' config map exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"),  kubernetes.ConfigMapName, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' config map doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ConfigMapName)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' config map exists", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.ConfigMapName)
	}

	if doesResourceExist, err := kubernetesProvider.DoesServiceAccountExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' service account exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"),  kubernetes.ServiceAccountName, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' service account doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ServiceAccountName)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' service account exists", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.ServiceAccountName)
	}

	if doesResourceExist, err := kubernetesProvider.DoesClusterRoleExist(ctx, kubernetes.ClusterRoleName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' cluster role exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"),  kubernetes.ClusterRoleName, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' cluster role doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ClusterRoleName)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' cluster role exists", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.ClusterRoleName)
	}

	if doesResourceExist, err := kubernetesProvider.DoesClusterRoleBindingExist(ctx, kubernetes.ClusterRoleBindingName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' cluster role binding exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"),  kubernetes.ClusterRoleBindingName, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' cluster role binding doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ClusterRoleBindingName)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' cluster role binding exists", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.ClusterRoleBindingName)
	}

	if doesResourceExist, err := kubernetesProvider.DoesRoleExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.DaemonRoleName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' role exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"),  kubernetes.DaemonRoleName, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' role doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.DaemonRoleName)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' role exists", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.DaemonRoleName)
	}

	if doesResourceExist, err := kubernetesProvider.DoesRoleBindingExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.DaemonRoleBindingName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' role binding exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"),  kubernetes.DaemonRoleBindingName, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' role binding doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.DaemonRoleBindingName)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' role binding exists", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.DaemonRoleBindingName)
	}

	if doesResourceExist, err := kubernetesProvider.DoesPersistentVolumeClaimExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.PersistentVolumeClaimName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' persistent volume claim exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"),  kubernetes.PersistentVolumeClaimName, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' persistent volume claim doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.PersistentVolumeClaimName)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' persistent volume claim exists", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.PersistentVolumeClaimName)
	}

	if doesResourceExist, err := kubernetesProvider.DoesDeploymentExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' deployment exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"),  kubernetes.ApiServerPodName, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' deployment doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ApiServerPodName)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' deployment exists", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.ApiServerPodName)
	}

	if doesResourceExist, err := kubernetesProvider.DoesServiceExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		logger.Log.Errorf("%v error checking if '%v' service exists, err: %v", fmt.Sprintf(uiUtils.Red, "✗"),  kubernetes.ApiServerPodName, err)
		allResourcesExists = false
	} else if !doesResourceExist {
		logger.Log.Errorf("%v '%v' service doesn't exist", fmt.Sprintf(uiUtils.Red, "✗"), kubernetes.ApiServerPodName)
		allResourcesExists = false
	} else {
		logger.Log.Infof("%v '%v' service exists", fmt.Sprintf(uiUtils.Green, "√"), kubernetes.ApiServerPodName)
	}

	return allResourcesExists
}
