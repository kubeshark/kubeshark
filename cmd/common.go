package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"regexp"
	"time"

	"github.com/kubeshark/kubeshark/apiserver"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/kubeshark/kubeshark/kubeshark/fsUtils"
	"github.com/kubeshark/kubeshark/resources"
	"github.com/kubeshark/kubeshark/uiUtils"
	"github.com/kubeshark/worker/models"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
)

func startProxyReportErrorIfAny(kubernetesProvider *kubernetes.Provider, ctx context.Context, cancel context.CancelFunc, serviceName string, srcPort uint16, dstPort uint16, healthCheck string) {
	httpServer, err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.ProxyHost, srcPort, dstPort, config.Config.KubesharkResourcesNamespace, serviceName, cancel)
	if err != nil {
		log.Printf(uiUtils.Error, fmt.Sprintf("Error occured while running k8s proxy %v\n"+
			"Try setting different port by using --%s", errormessage.FormatError(err), configStructs.GuiPortTapName))
		cancel()
		return
	}

	provider := apiserver.NewProvider(kubernetes.GetLocalhostOnPort(srcPort), apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := provider.TestConnection(healthCheck); err != nil {
		log.Printf("Couldn't connect using proxy, stopping proxy and trying to create port-forward")
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("Error occurred while stopping proxy %v", errormessage.FormatError(err))
		}

		podRegex, _ := regexp.Compile(kubernetes.ApiServerPodName)
		if _, err := kubernetes.NewPortForward(kubernetesProvider, config.Config.KubesharkResourcesNamespace, podRegex, srcPort, dstPort, ctx, cancel); err != nil {
			log.Printf(uiUtils.Error, fmt.Sprintf("Error occured while running port forward [%s] %v\n"+
				"Try setting different port by using --%s", podRegex, errormessage.FormatError(err), configStructs.GuiPortTapName))
			cancel()
			return
		}

		provider = apiserver.NewProvider(kubernetes.GetLocalhostOnPort(srcPort), apiserver.DefaultRetries, apiserver.DefaultTimeout)
		if err := provider.TestConnection(healthCheck); err != nil {
			log.Printf(uiUtils.Error, fmt.Sprintf("Couldn't connect to [%s].", serviceName))
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
		log.Printf("cannot establish http-proxy connection to the Kubernetes cluster. If you’re using Lens or similar tool, please run kubeshark with regular kubectl config using --%v %v=$HOME/.kube/config flag", config.SetCommandName, config.KubeConfigPathConfigName)
	} else {
		log.Print(err)
	}
}

func finishKubesharkExecution(kubernetesProvider *kubernetes.Provider, isNsRestrictedMode bool, kubesharkResourcesNamespace string) {
	removalCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()
	dumpLogsIfNeeded(removalCtx, kubernetesProvider)
	resources.CleanUpKubesharkResources(removalCtx, cancel, kubernetesProvider, isNsRestrictedMode, kubesharkResourcesNamespace)
}

func dumpLogsIfNeeded(ctx context.Context, kubernetesProvider *kubernetes.Provider) {
	if !config.Config.DumpLogs {
		return
	}
	kubesharkDir := kubeshark.GetKubesharkFolderPath()
	filePath := path.Join(kubesharkDir, fmt.Sprintf("kubeshark_logs_%s.zip", time.Now().Format("2006_01_02__15_04_05")))
	if err := fsUtils.DumpLogs(ctx, kubernetesProvider, filePath); err != nil {
		log.Printf("Failed dump logs %v", err)
	}
}

func getSerializedKubesharkAgentConfig(kubesharkAgentConfig *models.Config) (string, error) {
	serializedConfig, err := json.Marshal(kubesharkAgentConfig)
	if err != nil {
		return "", err
	}

	return string(serializedConfig), nil
}
