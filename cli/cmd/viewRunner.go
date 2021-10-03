package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/version"
	"github.com/up9inc/mizu/cli/uiUtils"
)

func runMizuView() {
	kubernetesProvider, err := kubernetes.NewProvider(config.Config.KubeConfigPath())
	if err != nil {
		logger.Log.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exists, err := kubernetesProvider.DoesServicesExist(ctx, config.Config.MizuResourcesNamespace, mizu.ApiServerPodName)
	if err != nil {
		logger.Log.Errorf("Failed to found mizu service %v", err)
		cancel()
		return
	}
	if !exists {
		logger.Log.Infof("%s service not found, you should run `mizu tap` command first", mizu.ApiServerPodName)
		cancel()
		return
	}

	url := GetApiServerUrl()

	response, err := http.Get(fmt.Sprintf("%s/", url))
	if err == nil && response.StatusCode == 200 {
		logger.Log.Infof("Found a running service %s and open port %d", mizu.ApiServerPodName, config.Config.View.GuiPort)
		return
	}
	logger.Log.Infof("Establishing connection to k8s cluster...")
	go startProxyReportErrorIfAny(kubernetesProvider, cancel)

	if err := apiserver.Provider.InitAndTestConnection(GetApiServerUrl()); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Couldn't connect to API server, for more info check logs at %s", logger.GetLogFilePath()))
		return
	}

	logger.Log.Infof("Mizu is available at %s\n", url)
	uiUtils.OpenBrowser(url)
	if isCompatible, err := version.CheckVersionCompatibility(); err != nil {
		logger.Log.Errorf("Failed to check versions compatibility %v", err)
		cancel()
		return
	} else if !isCompatible {
		cancel()
		return
	}

	waitForFinish(ctx, cancel)
}
