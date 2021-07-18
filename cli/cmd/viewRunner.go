package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"net/http"
)

func runMizuView(mizuViewOptions *MizuViewOptions) {
	kubernetesProvider := kubernetes.NewProvider("")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exists, err := kubernetesProvider.DoesServicesExist(ctx, mizu.ResourcesNamespace, mizu.AggregatorPodName)
	if err != nil {
		panic(err)
	}
	if !exists {
		fmt.Printf("The %s service not found\n", mizu.AggregatorPodName)
		return
	}

	mizuProxiedUrl := kubernetes.GetMizuCollectorProxiedHostAndPath(mizuViewOptions.GuiPort)
	_, err = http.Get(fmt.Sprintf("http://%s/", mizuProxiedUrl))
	if err == nil {
		fmt.Printf("Found a running service %s and open port %d\n", mizu.AggregatorPodName, mizuViewOptions.GuiPort)
		return
	}
	fmt.Printf("Found service %s, creating k8s proxy\n", mizu.AggregatorPodName)

	fmt.Printf("Mizu is available at  http://%s\n", kubernetes.GetMizuCollectorProxiedHostAndPath(mizuViewOptions.GuiPort))
	err = kubernetes.StartProxy(kubernetesProvider, mizuViewOptions.GuiPort, mizu.ResourcesNamespace, mizu.AggregatorPodName)
	if err != nil {
		fmt.Printf("Error occured while running k8s proxy %v\n", err)
	}
}
