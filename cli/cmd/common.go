package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"time"

	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/resources"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"

	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared/kubernetes"
)

func GetApiServerUrl(port uint16) string {
	return fmt.Sprintf("http://%s", kubernetes.GetMizuApiServerProxiedHostAndPath(port))
}

func startProxyReportErrorIfAny(kubernetesProvider *kubernetes.Provider, ctx context.Context, cancel context.CancelFunc, port uint16) {
	httpServer, err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.ProxyHost, port, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName, cancel)
	if err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error occured while running k8s proxy %v\n"+
			"Try setting different port by using --%s", errormessage.FormatError(err), configStructs.GuiPortTapName))
		cancel()
		return
	}

	provider := apiserver.NewProvider(GetApiServerUrl(port), apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := provider.TestConnection(); err != nil {
		logger.Log.Debugf("Couldn't connect using proxy, stopping proxy and trying to create port-forward")
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Log.Debugf("Error occurred while stopping proxy %v", errormessage.FormatError(err))
		}

		podRegex, _ := regexp.Compile(kubernetes.ApiServerPodName)
		if _, err := kubernetes.NewPortForward(kubernetesProvider, config.Config.MizuResourcesNamespace, podRegex, port, ctx, cancel); err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error occured while running port forward %v\n"+
				"Try setting different port by using --%s", errormessage.FormatError(err), configStructs.GuiPortTapName))
			cancel()
			return
		}

		provider = apiserver.NewProvider(GetApiServerUrl(port), apiserver.DefaultRetries, apiserver.DefaultTimeout)
		if err := provider.TestConnection(); err != nil {
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Couldn't connect to API server, for more info check logs at %s", fsUtils.GetLogFilePath()))
			cancel()
			return
		}
	}
}

func getKubernetesProviderForCli() (*kubernetes.Provider, error) {
	kubernetesProvider, err := kubernetes.NewProvider(config.Config.KubeConfigPath(), config.Config.KubeContext)
	if err != nil {
		handleKubernetesProviderError(err)
		return nil, err
	}

	if err := kubernetesProvider.ValidateNotProxy(); err != nil {
		handleKubernetesProviderError(err)
		return nil, err
	}

	kubernetesVersion, err := kubernetesProvider.GetKubernetesVersion()
	if err != nil {
		handleKubernetesProviderError(err)
		return nil, err
	}

	if err := kubernetes.ValidateKubernetesVersion(kubernetesVersion); err != nil {
		handleKubernetesProviderError(err)
		return nil, err
	}

	return kubernetesProvider, nil
}

func handleKubernetesProviderError(err error) {
	var clusterBehindProxyErr *kubernetes.ClusterBehindProxyError
	if ok := errors.As(err, &clusterBehindProxyErr); ok {
		logger.Log.Errorf("cannot establish http-proxy connection to the Kubernetes cluster. If you’re using Lens or similar tool, please run mizu with regular kubectl config using --%v %v=$HOME/.kube/config flag", config.SetCommandName, config.KubeConfigPathConfigName)
	} else {
		logger.Log.Error(err)
	}
}

func finishMizuExecution(kubernetesProvider *kubernetes.Provider, isNsRestrictedMode bool, mizuResourcesNamespace string) {
	removalCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()
	dumpLogsIfNeeded(removalCtx, kubernetesProvider)
	resources.CleanUpMizuResources(removalCtx, cancel, kubernetesProvider, isNsRestrictedMode, mizuResourcesNamespace)
}

func dumpLogsIfNeeded(ctx context.Context, kubernetesProvider *kubernetes.Provider) {
	if !config.Config.DumpLogs {
		return
	}
	mizuDir := mizu.GetMizuFolderPath()
	filePath := path.Join(mizuDir, fmt.Sprintf("mizu_logs_%s.zip", time.Now().Format("2006_01_02__15_04_05")))
	if err := fsUtils.DumpLogs(ctx, kubernetesProvider, filePath); err != nil {
		logger.Log.Errorf("Failed dump logs %v", err)
	}
}

func getSerializedMizuAgentConfig(mizuAgentConfig *shared.MizuAgentConfig) (string, error) {
	serializedConfig, err := json.Marshal(mizuAgentConfig)
	if err != nil {
		return "", err
	}

	return string(serializedConfig), nil
}
