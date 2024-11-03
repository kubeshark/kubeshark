package cmd

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/fsnotify/fsnotify"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
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

	kubernetesProvider, err := getKubernetesProviderForCli(false, false)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	watchScripts(kubernetesProvider, true)
}

func createScript(provider *kubernetes.Provider, script misc.ConfigMapScript) (index int64, err error) {
	var scripts map[int64]misc.ConfigMapScript
	scripts, err = kubernetes.ConfigGetScripts(provider)
	if err != nil {
		return
	}
	script.Active = kubernetes.IsActiveScript(provider, script.Title)
	index = int64(len(scripts))
	if script.Title != "New Script" {
		for i, v := range scripts {
			if v.Title == script.Title {
				index = int64(i)
			}
		}
	}
	scripts[index] = script

	log.Info().Str("title", script.Title).Bool("Active", script.Active).Int64("Index", index).Msg("Creating script")
	var data []byte
	data, err = json.Marshal(scripts)
	if err != nil {
		return
	}

	_, err = kubernetes.SetConfig(provider, kubernetes.CONFIG_SCRIPTING_SCRIPTS, string(data))
	if err != nil {
		return
	}

	return
}

func updateScript(provider *kubernetes.Provider, index int64, script misc.ConfigMapScript) (err error) {
	var scripts map[int64]misc.ConfigMapScript
	scripts, err = kubernetes.ConfigGetScripts(provider)
	if err != nil {
		return
	}
	script.Active = kubernetes.IsActiveScript(provider, script.Title)
	scripts[index] = script

	var data []byte
	data, err = json.Marshal(scripts)
	if err != nil {
		return
	}

	_, err = kubernetes.SetConfig(provider, kubernetes.CONFIG_SCRIPTING_SCRIPTS, string(data))
	if err != nil {
		return
	}

	return
}

func deleteScript(provider *kubernetes.Provider, index int64) (err error) {
	var scripts map[int64]misc.ConfigMapScript
	scripts, err = kubernetes.ConfigGetScripts(provider)
	if err != nil {
		return
	}
	err = kubernetes.DeleteActiveScriptByTitle(provider, scripts[index].Title)
	if err != nil {
		return
	}
	delete(scripts, index)

	var data []byte
	data, err = json.Marshal(scripts)
	if err != nil {
		return
	}

	_, err = kubernetes.SetConfig(provider, kubernetes.CONFIG_SCRIPTING_SCRIPTS, string(data))
	if err != nil {
		return
	}

	return
}

func watchScripts(provider *kubernetes.Provider, block bool) {
	files := make(map[string]int64)

	scripts, err := config.Config.Scripting.GetScripts()
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	for _, script := range scripts {
		index, err := createScript(provider, script.ConfigMap())
		if err != nil {
			log.Error().Err(err).Send()
			continue
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

	// Set up a context for graceful shutdown on Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C (SIGINT) to stop watching
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		log.Info().Msg("Received interrupt, stopping script watch.")
		cancel()        // Cancel the context to stop the watcher loop
		watcher.Close() // Close watcher explicitly to break out of for-select loop
	}()

	// Attempt to add the directory to the watcher
	if err := watcher.Add(config.Config.Scripting.Source); err != nil {
		log.Error().Err(err).Msg("Failed to add scripting source to watcher")
		return
	}

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				if !strings.HasSuffix(event.Name, "js") {
					log.Info().Str("file", event.Name).Msg("Ignoring file")
					continue
				}
				switch event.Op {
				case fsnotify.Create:
					script, err := misc.ReadScriptFile(event.Name)
					if err != nil {
						log.Error().Err(err).Send()
						continue
					}

					index, err := createScript(provider, script.ConfigMap())
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

					err = updateScript(provider, index, script.ConfigMap())
					if err != nil {
						log.Error().Err(err).Send()
						continue
					}

				case fsnotify.Rename:
					index := files[event.Name]
					err := deleteScript(provider, index)
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
				time.Sleep(5 * time.Second) // Retry after a delay
			}
		}
	}()

	if err := watcher.Add(config.Config.Scripting.Source); err != nil {
		log.Error().Err(err).Send()
	}

	log.Info().Str("directory", config.Config.Scripting.Source).Msg("Watching scripts against changes:")

	if block {
		<-ctx.Done()
	}
}
