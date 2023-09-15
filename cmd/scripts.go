package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/creasty/defaults"
	"github.com/fsnotify/fsnotify"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var scriptsCmd = &cobra.Command{
	Use:   "scripts",
	Short: "Watch the `scripting.source` directory for changes and update the scripts",
	RunE: func(cmd *cobra.Command, args []string) error {
		runScripts()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scriptsCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	scriptsCmd.Flags().Uint16(configStructs.ProxyFrontPortLabel, defaultTapConfig.Proxy.Front.Port, "Provide a custom port for the Kubeshark")
	scriptsCmd.Flags().String(configStructs.ProxyHostLabel, defaultTapConfig.Proxy.Host, "Provide a custom host for the Kubeshark")
	scriptsCmd.Flags().StringP(configStructs.ReleaseNamespaceLabel, "s", defaultTapConfig.Release.Namespace, "Release namespace of Kubeshark")
}

func runScripts() {
	if config.Config.Scripting.Source == "" {
		log.Error().Msg("`scripting.source` field is empty.")
		return
	}

	hubUrl := kubernetes.GetHubUrl()
	response, err := http.Get(fmt.Sprintf("%s/echo", hubUrl))
	if err != nil || response.StatusCode != 200 {
		log.Info().Msg(fmt.Sprintf(utils.Yellow, "Couldn't connect to Hub. Establishing proxy..."))
		runProxy(false, true)
	}

	connector = connect.NewConnector(kubernetes.GetHubUrl(), connect.DefaultRetries, connect.DefaultTimeout)

	watchScripts(true)
}

func watchScripts(block bool) {
	files := make(map[string]int64)

	scripts, err := config.Config.Scripting.GetScripts()
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	for _, script := range scripts {
		index, err := connector.PostScript(script)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		files[script.Path] = index
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	if block {
		defer watcher.Close()
	}

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				switch event.Op {
				case fsnotify.Create:
					script, err := misc.ReadScriptFile(event.Name)
					if err != nil {
						log.Error().Err(err).Send()
						continue
					}

					index, err := connector.PostScript(script)
					if err != nil {
						log.Error().Err(err).Send()
						continue
					}

					files[script.Path] = index

				case fsnotify.Write:
					index := files[event.Name]
					script, err := misc.ReadScriptFile(event.Name)
					if err != nil {
						log.Error().Err(err).Send()
						continue
					}

					err = connector.PutScript(script, index)
					if err != nil {
						log.Error().Err(err).Send()
						continue
					}

				case fsnotify.Rename:
					index := files[event.Name]
					err := connector.DeleteScript(index)
					if err != nil {
						log.Error().Err(err).Send()
						continue
					}

				default:
					// pass
				}

			// watch for errors
			case err := <-watcher.Errors:
				log.Error().Err(err).Send()
			}
		}
	}()

	if err := watcher.Add(config.Config.Scripting.Source); err != nil {
		log.Error().Err(err).Send()
	}

	log.Info().Str("directory", config.Config.Scripting.Source).Msg("Watching scripts against changes:")

	if block {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		utils.WaitForTermination(ctx, cancel)
	}
}
