package check

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/kubeshark/kubeshark/apiserver"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/uiUtils"
)

func ServerConnection(kubernetesProvider *kubernetes.Provider) bool {
	log.Printf("\nAPI-server-connectivity\n--------------------")

	serverUrl := kubernetes.GetLocalhostOnPort(config.Config.Hub.PortForward.SrcPort)

	apiServerProvider := apiserver.NewProvider(serverUrl, 1, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(""); err == nil {
		log.Printf("%v found Kubeshark server tunnel available and connected successfully to API server", fmt.Sprintf(uiUtils.Green, "√"))
		return true
	}

	connectedToApiServer := false

	if err := checkProxy(serverUrl, kubernetesProvider); err != nil {
		log.Printf("%v couldn't connect to API server using proxy, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
	} else {
		connectedToApiServer = true
		log.Printf("%v connected successfully to API server using proxy", fmt.Sprintf(uiUtils.Green, "√"))
	}

	if err := checkPortForward(serverUrl, kubernetesProvider); err != nil {
		log.Printf("%v couldn't connect to API server using port-forward, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
	} else {
		connectedToApiServer = true
		log.Printf("%v connected successfully to API server using port-forward", fmt.Sprintf(uiUtils.Green, "√"))
	}

	return connectedToApiServer
}

func checkProxy(serverUrl string, kubernetesProvider *kubernetes.Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.ProxyHost, config.Config.Hub.PortForward.SrcPort, config.Config.Hub.PortForward.DstPort, config.Config.KubesharkResourcesNamespace, kubernetes.ApiServerPodName, cancel)
	if err != nil {
		return err
	}

	apiServerProvider := apiserver.NewProvider(serverUrl, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(""); err != nil {
		return err
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error occurred while stopping proxy, err: %v", err)
	}

	return nil
}

func checkPortForward(serverUrl string, kubernetesProvider *kubernetes.Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podRegex, _ := regexp.Compile(kubernetes.ApiServerPodName)
	forwarder, err := kubernetes.NewPortForward(kubernetesProvider, config.Config.KubesharkResourcesNamespace, podRegex, config.Config.Tap.GuiPort, config.Config.Tap.GuiPort, ctx, cancel)
	if err != nil {
		return err
	}

	apiServerProvider := apiserver.NewProvider(serverUrl, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(""); err != nil {
		return err
	}

	forwarder.Close()

	return nil
}
