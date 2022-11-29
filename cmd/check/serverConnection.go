package check

import (
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/rs/zerolog/log"
)

func ServerConnection(kubernetesProvider *kubernetes.Provider) bool {
	log.Info().Str("procedure", "server-connectivity").Msg("Checking:")

	var connectedToHub, connectedToFront bool

	if err := checkProxy(kubernetes.GetLocalhostOnPort(config.Config.Hub.PortForward.SrcPort), "/echo", kubernetesProvider); err != nil {
		log.Error().Err(err).Msg("Couldn't connect to Hub using proxy!")
	} else {
		connectedToHub = true
		log.Info().Msg("Connected successfully to Hub using proxy.")
	}

	if err := checkProxy(kubernetes.GetLocalhostOnPort(config.Config.Front.PortForward.SrcPort), "", kubernetesProvider); err != nil {
		log.Error().Err(err).Msg("Couldn't connect to Front using proxy!")
	} else {
		connectedToFront = true
		log.Info().Msg("Connected successfully to Front using proxy.")
	}

	return connectedToHub && connectedToFront
}

func checkProxy(serverUrl string, path string, kubernetesProvider *kubernetes.Provider) error {
	log.Info().Str("url", serverUrl).Msg("Connecting:")
	connector := connect.NewConnector(serverUrl, connect.DefaultRetries, connect.DefaultTimeout)
	if err := connector.TestConnection(path); err != nil {
		return err
	}

	return nil
}
