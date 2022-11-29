package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
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
		exists, err := kubernetesProvider.DoesServiceExist(ctx, config.Config.ResourcesNamespace, kubernetes.HubServiceName)
		if err != nil {
			log.Error().
				Str("name", "kubeshark").
				Err(err).
				Msg("Failed to found service!")
			cancel()
			return
		}
		if !exists {
			log.Printf("%s service not found, you should run `kubeshark tap` command first", kubernetes.HubServiceName)
			log.Error().
				Str("name", kubernetes.HubServiceName).
				Str("tap-command", "kubeshark tap").
				Msg("Service not found! You should run the tap command first:")
			cancel()
			return
		}

		url = kubernetes.GetLocalhostOnPort(config.Config.Front.PortForward.SrcPort)

		response, err := http.Get(fmt.Sprintf("%s/", url))
		if err == nil && response.StatusCode == 200 {
			log.Info().
				Str("name", kubernetes.HubServiceName).
				Int("port", int(config.Config.Front.PortForward.SrcPort)).
				Msg("Found a running service.")
			return
		}
		log.Info().Msg("Establishing connection to k8s cluster...")
		startProxyReportErrorIfAny(kubernetesProvider, ctx, cancel, kubernetes.FrontServiceName, config.Config.Front.PortForward.SrcPort, config.Config.Front.PortForward.DstPort, "")
	}

	connector := connect.NewConnector(url, connect.DefaultRetries, connect.DefaultTimeout)
	if err := connector.TestConnection(""); err != nil {
		log.Error().Msg(fmt.Sprintf(utils.Red, "Couldn't connect to Hub."))
		return
	}

	log.Info().Msg(fmt.Sprintf("Kubeshark is available at %s", url))

	if !config.Config.HeadlessMode {
		utils.OpenBrowser(url)
	}

	utils.WaitForFinish(ctx, cancel)
}
