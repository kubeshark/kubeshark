package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kubeshark/kubeshark/cli/utils"

	"github.com/kubeshark/kubeshark/cli/apiserver"
	"github.com/kubeshark/kubeshark/cli/config"
	"github.com/kubeshark/kubeshark/cli/kubeshark/fsUtils"
	"github.com/kubeshark/kubeshark/cli/uiUtils"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/shared/kubernetes"
)

func runKubesharkView() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url := config.Config.View.Url

	if url == "" {
		exists, err := kubernetesProvider.DoesServiceExist(ctx, config.Config.KubesharkResourcesNamespace, kubernetes.ApiServerPodName)
		if err != nil {
			logger.Log.Errorf("Failed to found kubeshark service %v", err)
			cancel()
			return
		}
		if !exists {
			logger.Log.Infof("%s service not found, you should run `kubeshark tap` command first", kubernetes.ApiServerPodName)
			cancel()
			return
		}

		url = kubernetes.GetLocalhostOnPort(config.Config.Front.PortForward.SrcPort)

		response, err := http.Get(fmt.Sprintf("%s/", url))
		if err == nil && response.StatusCode == 200 {
			logger.Log.Infof("Found a running service %s and open port %d", kubernetes.ApiServerPodName, config.Config.Front.PortForward.SrcPort)
			return
		}
		logger.Log.Infof("Establishing connection to k8s cluster...")
		startProxyReportErrorIfAny(kubernetesProvider, ctx, cancel, "front", config.Config.Front.PortForward.SrcPort, config.Config.Front.PortForward.DstPort, "")
	}

	apiServerProvider := apiserver.NewProvider(url, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(""); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Couldn't connect to API server, for more info check logs at %s", fsUtils.GetLogFilePath()))
		return
	}

	logger.Log.Infof("Kubeshark is available at %s", url)

	if !config.Config.HeadlessMode {
		uiUtils.OpenBrowser(url)
	}

	utils.WaitForFinish(ctx, cancel)
}
