package check

import (
	"context"
	"regexp"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/rs/zerolog/log"
)

func ServerConnection(kubernetesProvider *kubernetes.Provider) bool {
	log.Info().Msg("[hub-connectivity]")

	serverUrl := kubernetes.GetLocalhostOnPort(config.Config.Hub.PortForward.SrcPort)

	connector := connect.NewConnector(serverUrl, 1, connect.DefaultTimeout)
	if err := connector.TestConnection(""); err == nil {
		log.Info().Msg("Found Kubeshark server tunnel available and connected successfully to Hub!")
		return true
	}

	connectedToHub := false

	if err := checkProxy(serverUrl, kubernetesProvider); err != nil {
		log.Error().Err(err).Msg("Couldn't connect to Hub using proxy!")
	} else {
		connectedToHub = true
		log.Info().Msg("Connected successfully to Hub using proxy.")
	}

	if err := checkPortForward(serverUrl, kubernetesProvider); err != nil {
		log.Error().Err(err).Msg("Couldn't connect to Hub using port-forward!")
	} else {
		connectedToHub = true
		log.Info().Msg("Connected successfully to Hub using port-forward.")
	}

	return connectedToHub
}

func checkProxy(serverUrl string, kubernetesProvider *kubernetes.Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.ProxyHost, config.Config.Hub.PortForward.SrcPort, config.Config.Hub.PortForward.DstPort, config.Config.ResourcesNamespace, kubernetes.HubServiceName, cancel)
	if err != nil {
		return err
	}

	connector := connect.NewConnector(serverUrl, connect.DefaultRetries, connect.DefaultTimeout)
	if err := connector.TestConnection(""); err != nil {
		return err
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("While stopping the proxy!")
	}

	return nil
}

func checkPortForward(serverUrl string, kubernetesProvider *kubernetes.Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podRegex, _ := regexp.Compile(kubernetes.HubPodName)
	forwarder, err := kubernetes.NewPortForward(kubernetesProvider, config.Config.ResourcesNamespace, podRegex, config.Config.Tap.GuiPort, config.Config.Tap.GuiPort, ctx, cancel)
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
