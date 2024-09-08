package cmd

import (
	"context"
	"fmt"

	"github.com/go-cmd/cmd"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

func runPprof() {
	runProxy(false, true)

	provider, err := getKubernetesProviderForCli(false, false)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hubPods, err := provider.ListPodsByAppLabel(ctx, config.Config.Tap.Release.Namespace, map[string]string{kubernetes.AppLabelKey: "hub"})
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to list hub pods!")
		cancel()
		return
	}

	workerPods, err := provider.ListPodsByAppLabel(ctx, config.Config.Tap.Release.Namespace, map[string]string{kubernetes.AppLabelKey: "worker"})
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to list worker pods!")
		cancel()
		return
	}

	fullscreen := true

	app := tview.NewApplication()
	list := tview.NewList()

	var currentCmd *cmd.Cmd

	i := 48
	for _, pod := range hubPods {
		for _, container := range pod.Spec.Containers {
			log.Info().Str("pod", pod.Name).Str("container", container.Name).Send()
			homeUrl := fmt.Sprintf("%s/debug/pprof/", kubernetes.GetHubUrl())
			modal := buildNewModal(
				pod,
				container,
				homeUrl,
				app,
				list,
				fullscreen,
				currentCmd,
			)
			list.AddItem(fmt.Sprintf("pod: %s container: %s", pod.Name, container.Name), pod.Spec.NodeName, rune(i), func() {
				app.SetRoot(modal, fullscreen)
			})
			i++
		}
	}

	for _, pod := range workerPods {
		for _, container := range pod.Spec.Containers {
			log.Info().Str("pod", pod.Name).Str("container", container.Name).Send()
			homeUrl := fmt.Sprintf("%s/pprof/%s/%s/", kubernetes.GetHubUrl(), pod.Status.HostIP, container.Name)
			modal := buildNewModal(
				pod,
				container,
				homeUrl,
				app,
				list,
				fullscreen,
				currentCmd,
			)
			list.AddItem(fmt.Sprintf("pod: %s container: %s", pod.Name, container.Name), pod.Spec.NodeName, rune(i), func() {
				app.SetRoot(modal, fullscreen)
			})
			i++
		}
	}

	list.AddItem("Quit", "Press to exit", 'q', func() {
		if currentCmd != nil {
			err = currentCmd.Stop()
			if err != nil {
				log.Error().Err(err).Str("name", currentCmd.Name).Msg("Failed to stop process!")
			}
		}
		app.Stop()
	})

	if err := app.SetRoot(list, fullscreen).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func buildNewModal(
	pod v1.Pod,
	container v1.Container,
	homeUrl string,
	app *tview.Application,
	list *tview.List,
	fullscreen bool,
	currentCmd *cmd.Cmd,
) *tview.Modal {
	return tview.NewModal().
		SetText(fmt.Sprintf("pod: %s container: %s", pod.Name, container.Name)).
		AddButtons([]string{
			"Open Debug Home Page",
			"Profile: CPU",
			"Profile: Memory",
			"Profile: Goroutine",
			"Cancel",
		}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			var err error
			port := fmt.Sprintf(":%d", config.Config.Tap.Pprof.Port)
			view := fmt.Sprintf("http://localhost%s/ui/%s", port, config.Config.Tap.Pprof.View)

			switch buttonLabel {
			case "Open Debug Home Page":
				utils.OpenBrowser(homeUrl)
			case "Profile: CPU":
				if currentCmd != nil {
					err = currentCmd.Stop()
					if err != nil {
						log.Error().Err(err).Str("name", currentCmd.Name).Msg("Failed to stop process!")
					}
				}
				currentCmd = cmd.NewCmd("go", "tool", "pprof", "-http", port, "-no_browser", fmt.Sprintf("%sprofile", homeUrl))
				currentCmd.Start()
				utils.OpenBrowser(view)
			case "Profile: Memory":
				if currentCmd != nil {
					err = currentCmd.Stop()
					if err != nil {
						log.Error().Err(err).Str("name", currentCmd.Name).Msg("Failed to stop process!")
					}
				}
				currentCmd = cmd.NewCmd("go", "tool", "pprof", "-http", port, "-no_browser", fmt.Sprintf("%sheap", homeUrl))
				currentCmd.Start()
				utils.OpenBrowser(view)
			case "Profile: Goroutine":
				if currentCmd != nil {
					err = currentCmd.Stop()
					if err != nil {
						log.Error().Err(err).Str("name", currentCmd.Name).Msg("Failed to stop process!")
					}
				}
				currentCmd = cmd.NewCmd("go", "tool", "pprof", "-http", port, "-no_browser", fmt.Sprintf("%sgoroutine", homeUrl))
				currentCmd.Start()
				utils.OpenBrowser(view)
			case "Cancel":
				if currentCmd != nil {
					err = currentCmd.Stop()
					if err != nil {
						log.Error().Err(err).Str("name", currentCmd.Name).Msg("Failed to stop process!")
					}
				}
				fallthrough
			default:
				app.SetRoot(list, fullscreen)
			}
		})
}
