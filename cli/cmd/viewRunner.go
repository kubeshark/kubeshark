package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"net/http"
)

func runMizuView() {
	kubernetesProvider, err := kubernetes.NewProvider(mizu.Config.View.KubeConfigPath)
	if err != nil {
		mizu.Log.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exists, err := kubernetesProvider.DoesServicesExist(ctx, mizu.Config.MizuResourcesNamespace, mizu.ApiServerPodName)
	if err != nil {
		mizu.Log.Errorf("Failed to found mizu service %v", err)
		cancel()
		return
	}
	if !exists {
		mizu.Log.Infof("%s service not found, you should run `mizu tap` command first", mizu.ApiServerPodName)
		cancel()
		return
	}

	mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.View.GuiPort)
	_, err = http.Get(fmt.Sprintf("http://%s/", mizuProxiedUrl))
	if err == nil {
		mizu.Log.Infof("Found a running service %s and open port %d", mizu.ApiServerPodName, mizu.Config.View.GuiPort)
		return
	}
	mizu.Log.Debugf("Found service %s, creating k8s proxy", mizu.ApiServerPodName)

	go kubernetes.StartProxyReportErrorIfAny(kubernetesProvider, cancel)

	mizu.Log.Infof("Mizu is available at  http://%s\n", kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.View.GuiPort))
	if isCompatible, err := mizu.CheckVersionCompatibility(mizu.Config.View.GuiPort); err != nil {
		mizu.Log.Errorf("Failed to check versions compatibility %v", err)
		cancel()
	} else if !isCompatible {
		mizu.Log.Errorf("Mizu Cli and Mizu server not same version")
		cancel()
	}

	waitForFinish(ctx, cancel)
}
