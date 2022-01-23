package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/utils"
	"net/http"

	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/mizu/version"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
)

func runMizuView() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url := config.Config.View.Url

	if url == "" {
		exists, err := kubernetesProvider.DoesServicesExist(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName)
		if err != nil {
			logger.Log.Errorf("Failed to found mizu service %v", err)
			cancel()
			return
		}
		if !exists {
			logger.Log.Infof("%s service not found, you should run `mizu tap` command first", kubernetes.ApiServerPodName)
			cancel()
			return
		}

		url = GetApiServerUrl()

		response, err := http.Get(fmt.Sprintf("%s/", url))
		if err == nil && response.StatusCode == 200 {
			logger.Log.Infof("Found a running service %s and open port %d", kubernetes.ApiServerPodName, config.Config.View.GuiPort)
			return
		}
		logger.Log.Infof("Establishing connection to k8s cluster...")
		startProxyReportErrorIfAny(kubernetesProvider, cancel)
	}

	apiServerProvider := apiserver.NewProvider(url, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Couldn't connect to API server, for more info check logs at %s", fsUtils.GetLogFilePath()))
		return
	}

	logger.Log.Infof("Mizu is available at %s", url)

	if !config.Config.HeadlessMode {
		uiUtils.OpenBrowser(url)
	}

	if isCompatible, err := version.CheckVersionCompatibility(apiServerProvider); err != nil {
		logger.Log.Errorf("Failed to check versions compatibility %v", err)
		cancel()
		return
	} else if !isCompatible {
		cancel()
		return
	}

	utils.WaitForFinish(ctx, cancel)
}
