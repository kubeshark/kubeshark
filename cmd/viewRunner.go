package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/utils"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
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
		exists, err := kubernetesProvider.DoesServiceExist(ctx, config.Config.KubesharkResourcesNamespace, kubernetes.ApiServerServiceName)
		if err != nil {
			log.Printf("Failed to found kubeshark service %v", err)
			cancel()
			return
		}
		if !exists {
			log.Printf("%s service not found, you should run `kubeshark tap` command first", kubernetes.ApiServerServiceName)
			cancel()
			return
		}

		url = kubernetes.GetLocalhostOnPort(config.Config.Front.PortForward.SrcPort)

		response, err := http.Get(fmt.Sprintf("%s/", url))
		if err == nil && response.StatusCode == 200 {
			log.Printf("Found a running service %s and open port %d", kubernetes.ApiServerServiceName, config.Config.Front.PortForward.SrcPort)
			return
		}
		log.Printf("Establishing connection to k8s cluster...")
		startProxyReportErrorIfAny(kubernetesProvider, ctx, cancel, kubernetes.FrontServiceName, config.Config.Front.PortForward.SrcPort, config.Config.Front.PortForward.DstPort, "")
	}

	connector := connect.NewConnector(url, connect.DefaultRetries, connect.DefaultTimeout)
	if err := connector.TestConnection(""); err != nil {
		log.Printf(utils.Error, "Couldn't connect to API server.")
		return
	}

	log.Printf("Kubeshark is available at %s", url)

	if !config.Config.HeadlessMode {
		utils.OpenBrowser(url)
	}

	utils.WaitForFinish(ctx, cancel)
}
