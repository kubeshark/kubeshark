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

func runProxy(block bool, noBrowser bool) {
	kubernetesProvider, err := getKubernetesProviderForCli(false, false)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exists, err := kubernetesProvider.DoesServiceExist(ctx, config.Config.Tap.ReleaseNamespace, kubernetes.FrontServiceName)
	if err != nil {
		log.Error().
			Str("service", kubernetes.FrontServiceName).
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

	exists, err = kubernetesProvider.DoesServiceExist(ctx, config.Config.Tap.ReleaseNamespace, kubernetes.HubServiceName)
	if err != nil {
		log.Error().
			Str("service", kubernetes.HubServiceName).
			Err(err).
			Msg("Failed to found service!")
		cancel()
		return
	}

	if !exists {
		log.Error().
			Str("service", kubernetes.HubServiceName).
			Str("command", fmt.Sprintf("%s %s", misc.Program, tapCmd.Use)).
			Msg("Service not found! You should run the command first:")
		cancel()
		return
	}

	var establishedProxy bool

	hubUrl := kubernetes.GetProxyOnPort(config.Config.Tap.Proxy.Hub.Port)
	response, err := http.Get(fmt.Sprintf("%s/echo", hubUrl))
	if err == nil && response.StatusCode == 200 {
		log.Info().
			Str("service", kubernetes.HubServiceName).
			Int("port", int(config.Config.Tap.Proxy.Hub.Port)).
			Msg("Found a running service.")

		okToOpen("Hub", hubUrl, true)
	} else {
		startProxyReportErrorIfAny(
			kubernetesProvider,
			ctx,
			kubernetes.HubServiceName,
			kubernetes.HubPodName,
			configStructs.ProxyHubPortLabel,
			config.Config.Tap.Proxy.Hub.Port,
			configStructs.ContainerPort,
			"/echo",
		)
		connector := connect.NewConnector(hubUrl, connect.DefaultRetries, connect.DefaultTimeout)
		if err := connector.TestConnection("/echo"); err != nil {
			log.Error().Msg(fmt.Sprintf(utils.Red, "Couldn't connect to Hub."))
			return
		}

		establishedProxy = true
		okToOpen("Hub", hubUrl, true)
	}

	frontUrl := kubernetes.GetProxyOnPort(config.Config.Tap.Proxy.Front.Port)
	response, err = http.Get(fmt.Sprintf("%s/", frontUrl))
	if err == nil && response.StatusCode == 200 {
		log.Info().
			Str("service", kubernetes.FrontServiceName).
			Int("port", int(config.Config.Tap.Proxy.Front.Port)).
			Msg("Found a running service.")

		okToOpen("Kubeshark", frontUrl, noBrowser)
	} else {
		startProxyReportErrorIfAny(
			kubernetesProvider,
			ctx,
			kubernetes.FrontServiceName,
			kubernetes.FrontPodName,
			configStructs.ProxyFrontPortLabel,
			config.Config.Tap.Proxy.Front.Port,
			configStructs.ContainerPort,
			"",
		)
		connector := connect.NewConnector(frontUrl, connect.DefaultRetries, connect.DefaultTimeout)
		if err := connector.TestConnection(""); err != nil {
			log.Error().Msg(fmt.Sprintf(utils.Red, "Couldn't connect to Front."))
			return
		}

		establishedProxy = true
		okToOpen("Kubeshark", frontUrl, noBrowser)
	}

	if establishedProxy && block {
		utils.WaitForTermination(ctx, cancel)
	}
}

func okToOpen(name string, url string, noBrowser bool) {
	log.Info().Str("url", url).Msg(fmt.Sprintf(utils.Green, fmt.Sprintf("%s is available at:", name)))

	if !config.Config.HeadlessMode && !noBrowser {
		utils.OpenBrowser(url)
	}
}
