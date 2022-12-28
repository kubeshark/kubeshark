package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"time"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/kubeshark"
	"github.com/kubeshark/kubeshark/kubeshark/fsUtils"
	"github.com/kubeshark/kubeshark/resources"
	"github.com/rs/zerolog/log"
)

func startProxyReportErrorIfAny(kubernetesProvider *kubernetes.Provider, ctx context.Context, cancel context.CancelFunc, serviceName string, proxyPortLabel string, srcPort uint16, dstPort uint16, healthCheck string) {
	httpServer, err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.Proxy.Host, srcPort, config.Config.ResourcesNamespace, serviceName, cancel)
	if err != nil {
		log.Error().
			Err(errormessage.FormatError(err)).
			Msg(fmt.Sprintf("Error occured while running k8s proxy. Try setting different port by using --%s", proxyPortLabel))
		cancel()
		return
	}

	connector := connect.NewConnector(kubernetes.GetLocalhostOnPort(srcPort), connect.DefaultRetries, connect.DefaultTimeout)
	if err := connector.TestConnection(healthCheck); err != nil {
		log.Error().Msg("Couldn't connect using proxy, stopping proxy and trying to create port-forward..")
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Error().
				Err(errormessage.FormatError(err)).
				Msg("Error occurred while stopping proxy.")
		}

		podRegex, _ := regexp.Compile(kubernetes.HubPodName)
		if _, err := kubernetes.NewPortForward(kubernetesProvider, config.Config.ResourcesNamespace, podRegex, srcPort, dstPort, ctx, cancel); err != nil {
			log.Error().
				Str("pod-regex", podRegex.String()).
				Err(errormessage.FormatError(err)).
				Msg(fmt.Sprintf("Error occured while running port forward. Try setting different port by using --%s", proxyPortLabel))
			cancel()
			return
		}

		connector = connect.NewConnector(kubernetes.GetLocalhostOnPort(srcPort), connect.DefaultRetries, connect.DefaultTimeout)
		if err := connector.TestConnection(healthCheck); err != nil {
			log.Error().
				Str("service", serviceName).
				Err(errormessage.FormatError(err)).
				Msg("Couldn't connect to service.")
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
		log.Error().Msg(fmt.Sprintf("Cannot establish http-proxy connection to the Kubernetes cluster. If you’re using Lens or similar tool, please run kubeshark with regular kubectl config using --%v %v=$HOME/.kube/config flag", config.SetCommandName, config.KubeConfigPathConfigName))
	} else {
		log.Error().Err(err).Send()
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
		log.Error().Err(err).Msg("Failed to dump logs.")
	}
}

func getSerializedTapConfig(conf *models.Config) (string, error) {
	serializedConfig, err := json.Marshal(conf)
	if err != nil {
		return "", err
	}

	return string(serializedConfig), nil
}
