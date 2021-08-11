package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/version"
	"net/http"
)

func runMizuView() {
	kubernetesProvider, err := kubernetes.NewProvider(config.Config.View.KubeConfigPath)
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

	mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(config.Config.View.GuiPort)
	_, err = http.Get(fmt.Sprintf("http://%s/", mizuProxiedUrl))
	if err == nil {
		logger.Log.Infof("Found a running service %s and open port %d", mizu.ApiServerPodName, config.Config.View.GuiPort)
		return
	}
	logger.Log.Debugf("Found service %s, creating k8s proxy", mizu.ApiServerPodName)

	go startProxyReportErrorIfAny(kubernetesProvider, cancel)

	logger.Log.Infof("Mizu is available at  http://%s\n", kubernetes.GetMizuApiServerProxiedHostAndPath(config.Config.View.GuiPort))
	if isCompatible, err := version.CheckVersionCompatibility(config.Config.View.GuiPort); err != nil {
		logger.Log.Errorf("Failed to check versions compatibility %v", err)
		cancel()
		return
	} else if !isCompatible {
		cancel()
		return
	}

	waitForFinish(ctx, cancel)
}
