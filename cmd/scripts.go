package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/creasty/defaults"
	"github.com/fsnotify/fsnotify"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var scriptsCmd = &cobra.Command{
	Use:   "scripts",
	Short: "Watch the `scripting.source` and/or `scripting.sources` folders for changes and update the scripts",
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
	if config.Config.Scripting.Source == "" && len(config.Config.Scripting.Sources) == 0 {
		log.Error().Msg("Both `scripting.source` and `scripting.sources` fields are empty.")
		return
	}

	kubernetesProvider, err := getKubernetesProviderForCli(false, false)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	wg.Add(1)
	go func() {
		defer wg.Done()
		watchConfigMap(ctx, kubernetesProvider)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		watchScripts(ctx, kubernetesProvider, true)
	}()

	go func() {
		<-signalChan
		log.Debug().Msg("Received interrupt, stopping watchers.")
		cancel()
	}()

	wg.Wait()

}

func createScript(provider *kubernetes.Provider, script misc.ConfigMapScript) (index int64, err error) {
	const maxRetries = 5
	var scripts map[int64]misc.ConfigMapScript

	for i := 0; i < maxRetries; i++ {
		scripts, err = kubernetes.ConfigGetScripts(provider)
		if err != nil {
			return
		}
		script.Active = kubernetes.IsActiveScript(provider, script.Title)
		index = 0
		if script.Title != "New Script" {
			for i, v := range scripts {
				if index <= i {
					index = i + 1
				}
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
		if err == nil {
			return index, nil
		}

		if k8serrors.IsConflict(err) {
			log.Warn().Err(err).Msg("Conflict detected, retrying update...")
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return 0, err
	}

	log.Error().Msg("Max retries reached for creating script due to conflicts.")
	return 0, errors.New("max retries reached due to conflicts while creating script")
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

func watchScripts(ctx context.Context, provider *kubernetes.Provider, block bool) {
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		log.Debug().Msg("Received interrupt, stopping script watch.")
		cancel()
		watcher.Close()
	}()

	if err := watcher.Add(config.Config.Scripting.Source); err != nil {
		log.Error().Err(err).Msg("Failed to add scripting source to watcher")
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Script watcher exiting gracefully.")
				return

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

			case err, ok := <-watcher.Errors:
				if !ok {
					log.Info().Msg("Watcher errors channel closed.")
					return
				}
				log.Error().Err(err).Msg("Watcher error encountered")
			}
		}
	}()

	if err := watcher.Add(config.Config.Scripting.Source); err != nil {
		log.Error().Err(err).Send()
	}

	for _, source := range config.Config.Scripting.Sources {
		if err := watcher.Add(source); err != nil {
			log.Error().Err(err).Send()
		}
	}

	log.Info().Str("folder", config.Config.Scripting.Source).Interface("folders", config.Config.Scripting.Sources).Msg("Watching scripts against changes:")

	if block {
		<-ctx.Done()
	}
}

func watchConfigMap(ctx context.Context, provider *kubernetes.Provider) {
	clientset := provider.GetClientSet()
	configMapName := kubernetes.SELF_RESOURCES_PREFIX + kubernetes.SUFFIX_CONFIG_MAP

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("ConfigMap watcher exiting gracefully.")
			return

		default:
			watcher, err := clientset.CoreV1().ConfigMaps(config.Config.Tap.Release.Namespace).Watch(context.TODO(), metav1.ListOptions{
				FieldSelector: "metadata.name=" + configMapName,
			})
			if err != nil {
				log.Warn().Err(err).Msg("ConfigMap not found, retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
				continue
			}

			for event := range watcher.ResultChan() {
				select {
				case <-ctx.Done():
					log.Info().Msg("ConfigMap watcher loop exiting gracefully.")
					watcher.Stop()
					return

				default:
					if event.Type == watch.Added {
						log.Info().Msg("ConfigMap created or modified")
						runScriptsSync(provider)
					} else if event.Type == watch.Deleted {
						log.Warn().Msg("ConfigMap deleted, waiting for recreation...")
						watcher.Stop()
						break
					}
				}
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func runScriptsSync(provider *kubernetes.Provider) {
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
	log.Info().Msg("Synchronized scripts with ConfigMap.")
}
