package check

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/utils"
)

func ServerConnection(kubernetesProvider *kubernetes.Provider) bool {
	log.Printf("\nAPI-server-connectivity\n--------------------")

	serverUrl := kubernetes.GetLocalhostOnPort(config.Config.Hub.PortForward.SrcPort)

	connector := connect.NewConnector(serverUrl, 1, connect.DefaultTimeout)
	if err := connector.TestConnection(""); err == nil {
		log.Printf("%v found Kubeshark server tunnel available and connected successfully to API server", fmt.Sprintf(utils.Green, "√"))
		return true
	}

	connectedToApiServer := false

	if err := checkProxy(serverUrl, kubernetesProvider); err != nil {
		log.Printf("%v couldn't connect to API server using proxy, err: %v", fmt.Sprintf(utils.Red, "✗"), err)
	} else {
		connectedToApiServer = true
		log.Printf("%v connected successfully to API server using proxy", fmt.Sprintf(utils.Green, "√"))
	}

	if err := checkPortForward(serverUrl, kubernetesProvider); err != nil {
		log.Printf("%v couldn't connect to API server using port-forward, err: %v", fmt.Sprintf(utils.Red, "✗"), err)
	} else {
		connectedToApiServer = true
		log.Printf("%v connected successfully to API server using port-forward", fmt.Sprintf(utils.Green, "√"))
	}

	return connectedToApiServer
}

func checkProxy(serverUrl string, kubernetesProvider *kubernetes.Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.ProxyHost, config.Config.Hub.PortForward.SrcPort, config.Config.Hub.PortForward.DstPort, config.Config.KubesharkResourcesNamespace, kubernetes.ApiServerServiceName, cancel)
	if err != nil {
		return err
	}

	connector := connect.NewConnector(serverUrl, connect.DefaultRetries, connect.DefaultTimeout)
	if err := connector.TestConnection(""); err != nil {
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

	connector := connect.NewConnector(serverUrl, connect.DefaultRetries, connect.DefaultTimeout)
	if err := connector.TestConnection(""); err != nil {
		return err
	}

	forwarder.Close()

	return nil
}
