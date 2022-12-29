package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
)

func runProxy() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exists, err := kubernetesProvider.DoesServiceExist(ctx, config.Config.SelfNamespace, kubernetes.FrontServiceName)
	if err != nil {
		log.Error().
			Str("service", misc.Program).
			Err(err).
			Msg("Failed to found service!")
		cancel()
		return
	}

	if !exists {
		log.Error().
			Str("service", kubernetes.FrontServiceName).
			Str("command", fmt.Sprintf("%s %s", misc.Program, tapCmd.Use)).
			Msg("Service not found! You should run the command first:")
		cancel()
		return
	}

	url := kubernetes.GetLocalhostOnPort(config.Config.Tap.Proxy.Front.SrcPort)

	response, err := http.Get(fmt.Sprintf("%s/", url))
	if err == nil && response.StatusCode == 200 {
		log.Info().
			Str("service", kubernetes.FrontServiceName).
			Int("port", int(config.Config.Tap.Proxy.Front.SrcPort)).
			Msg("Found a running service.")

		okToOpen(url)
		return
	}
	log.Info().Msg("Establishing connection to K8s cluster...")
	startProxyReportErrorIfAny(kubernetesProvider, ctx, cancel, kubernetes.FrontServiceName, configStructs.ProxyFrontPortLabel, config.Config.Tap.Proxy.Front.SrcPort, config.Config.Tap.Proxy.Front.DstPort, "")

	connector := connect.NewConnector(url, connect.DefaultRetries, connect.DefaultTimeout)
	if err := connector.TestConnection(""); err != nil {
		log.Error().Msg(fmt.Sprintf(utils.Red, "Couldn't connect to Front."))
		return
	}

	okToOpen(url)

	utils.WaitForTermination(ctx, cancel)
}

func okToOpen(url string) {
	log.Info().Str("url", url).Msg(fmt.Sprintf(utils.Green, fmt.Sprintf("%s is available at:", misc.Software)))

	if !config.Config.HeadlessMode {
		utils.OpenBrowser(url)
	}
}
