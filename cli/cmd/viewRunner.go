package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/mizu"
	"net/http"
)

func runMizuView() {
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

	_, err = http.Get("http://localhost:8899/")
	if err == nil {
		fmt.Printf("Found a running service %s and open port 8899\n", mizu.AggregatorPodName)
		return
	}
	fmt.Printf("Found service %s, creating port forwarding to 8899\n", mizu.AggregatorPodName)
	portForwardApiPod(ctx, kubernetesProvider, cancel, &MizuTapOptions{GuiPort: 8899, MizuPodPort: 8899})
}
