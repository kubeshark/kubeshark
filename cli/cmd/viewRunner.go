package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/uiUtils"
	"k8s.io/client-go/tools/clientcmd"
)

func runMizuView() {
	kubernetesProvider, err := kubernetes.NewProvider(mizu.Config.View.KubeConfigPath)
	if err != nil {
		if clientcmd.IsEmptyConfig(err) {
			mizu.Log.Infof("Couldn't find the kube config file, or file is empty. Try adding '--kube-config=<path to kube config file>'")
			return
		}
		if clientcmd.IsConfigurationInvalid(err) {
			mizu.Log.Infof(uiUtils.Red, "Invalid kube config file. Try using a different config with '--kube-config=<path to kube config file>'")
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exists, err := kubernetesProvider.DoesServicesExist(ctx, mizu.Config.ResourcesNamespace(), mizu.ApiServerPodName)
	if err != nil {
		panic(err)
	}
	if !exists {
		mizu.Log.Infof("The %s service not found", mizu.ApiServerPodName)
		return
	}

	mizuProxiedUrl := kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.View.GuiPort)
	_, err = http.Get(fmt.Sprintf("http://%s/", mizuProxiedUrl))
	if err == nil {
		mizu.Log.Infof("Found a running service %s and open port %d", mizu.ApiServerPodName, mizu.Config.View.GuiPort)
		return
	}
	mizu.Log.Infof("Found service %s, creating k8s proxy", mizu.ApiServerPodName)

	mizu.Log.Infof("Mizu is available at  http://%s\n", kubernetes.GetMizuApiServerProxiedHostAndPath(mizu.Config.View.GuiPort))
	err = kubernetes.StartProxy(kubernetesProvider, mizu.Config.View.GuiPort, mizu.Config.ResourcesNamespace(), mizu.ApiServerPodName)
	if err != nil {
		mizu.Log.Infof("Error occured while running k8s proxy %v", err)
	}
}
