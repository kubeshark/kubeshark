package cmd

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/kubeshark/kubeshark/docker"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/resources"
	"github.com/kubeshark/kubeshark/utils"

	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/rs/zerolog/log"
)

const cleanupTimeout = time.Minute

type tapState struct {
	startTime                time.Time
	targetNamespaces         []string
	selfServiceAccountExists bool
}

var state tapState
var connector *connect.Connector

type Readiness struct {
	Hub   bool
	Front bool
	Proxy bool
	sync.Mutex
}

var ready *Readiness

func tap() {
	ready = &Readiness{}
	state.startTime = time.Now()
	docker.SetRegistry(config.Config.Tap.Docker.Registry)
	docker.SetTag(config.Config.Tap.Docker.Tag)
	log.Info().Str("registry", docker.GetRegistry()).Str("tag", docker.GetTag()).Msg("Using Docker:")
	if config.Config.Tap.Pcap != "" {
		pcap(config.Config.Tap.Pcap)
		return
	}

	log.Info().
		Str("limit", config.Config.Tap.StorageLimit).
		Msg(fmt.Sprintf("%s will store the traffic up to a limit (per node). Oldest TCP/UDP streams will be removed once the limit is reached.", misc.Software))

	connector = connect.NewConnector(kubernetes.GetLocalhostOnPort(config.Config.Tap.Proxy.Hub.SrcPort), connect.DefaultRetries, connect.DefaultTimeout)

	kubernetesProvider, err := getKubernetesProviderForCli(false, false)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	state.targetNamespaces = kubernetesProvider.GetNamespaces()

	if config.Config.IsNsRestrictedMode() {
		if len(state.targetNamespaces) != 1 || !utils.Contains(state.targetNamespaces, config.Config.Tap.SelfNamespace) {
			log.Error().Msg(fmt.Sprintf("%s can't resolve IPs in other namespaces when running in namespace restricted mode. You can use the same namespace for --%s and --%s", misc.Software, configStructs.NamespacesLabel, configStructs.SelfNamespaceLabel))
			return
		}
	}

	log.Info().Strs("namespaces", state.targetNamespaces).Msg("Targeting pods in:")

	if err := printTargetedPodsPreview(ctx, kubernetesProvider, state.targetNamespaces); err != nil {
		log.Error().Err(errormessage.FormatError(err)).Msg("Error listing pods!")
	}

	if config.Config.Tap.DryRun {
		return
	}

	log.Info().Msg(fmt.Sprintf("Waiting for the creation of %s resources...", misc.Software))
	if state.selfServiceAccountExists, err = resources.CreateHubResources(ctx, kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.Tap.SelfNamespace, config.Config.Tap.Resources.Hub, config.Config.ImagePullPolicy(), config.Config.ImagePullSecrets(), config.Config.Tap.Debug); err != nil {
		var statusError *k8serrors.StatusError
		if errors.As(err, &statusError) && (statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists) {
			log.Info().Msg(fmt.Sprintf("%s is already running in this namespace, change the `selfnamespace` configuration or run `%s clean` to remove the currently running %s instance.", misc.Software, misc.Program, misc.Software))
			postHubStarted(ctx, kubernetesProvider, cancel, true)
			log.Info().Msg("Updated Hub about the changes in the config. Exiting.")
			printProxyCommandSuggestion()
		} else {
			defer resources.CleanUpSelfResources(ctx, cancel, kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.Tap.SelfNamespace)
			log.Error().Err(errormessage.FormatError(err)).Msg("Error creating resources!")
		}

		return
	}

	defer finishTapExecution(kubernetesProvider)

	go watchHubEvents(ctx, kubernetesProvider, cancel)
	go watchHubPod(ctx, kubernetesProvider, cancel)
	go watchFrontPod(ctx, kubernetesProvider, cancel)

	// block until exit signal or error
	utils.WaitForTermination(ctx, cancel)
	printProxyCommandSuggestion()
}

func printProxyCommandSuggestion() {
	log.Warn().
		Str("command", fmt.Sprintf("%s proxy", misc.Program)).
		Msg(fmt.Sprintf(utils.Yellow, "To re-establish a proxy/port-forward, run:"))
}

func finishTapExecution(kubernetesProvider *kubernetes.Provider) {
	finishSelfExecution(kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.Tap.SelfNamespace, true)
}

/*
This function is a bit problematic as it might be detached from the actual pods the Kubeshark that targets.
The alternative would be to wait for Hub to be ready and then query it for the pods it listens to, this has
the arguably worse drawback of taking a relatively very long time before the user sees which pods are targeted, if any.
*/
func printTargetedPodsPreview(ctx context.Context, kubernetesProvider *kubernetes.Provider, namespaces []string) error {
	if matchingPods, err := kubernetesProvider.ListAllRunningPodsMatchingRegex(ctx, config.Config.Tap.PodRegex(), namespaces); err != nil {
		return err
	} else {
		if len(matchingPods) == 0 {
			printNoPodsFoundSuggestion(namespaces)
		}
		for _, targetedPod := range matchingPods {
			log.Info().Msg(fmt.Sprintf("Targeted pod: %s", fmt.Sprintf(utils.Green, targetedPod.Name)))
		}
		return nil
	}
}

func printNoPodsFoundSuggestion(targetNamespaces []string) {
	var suggestionStr string
	if !utils.Contains(targetNamespaces, kubernetes.K8sAllNamespaces) {
		suggestionStr = ". You can also try selecting a different namespace with -n or target all namespaces with -A"
	}
	log.Warn().Msg(fmt.Sprintf("Did not find any currently running pods that match the regex argument, %s will automatically target matching pods if any are created later%s", misc.Software, suggestionStr))
}

func watchHubPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", kubernetes.HubPodName))
	podWatchHelper := kubernetes.NewPodWatchHelper(kubernetesProvider, podExactRegex)
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, podWatchHelper, []string{config.Config.Tap.SelfNamespace}, podWatchHelper)
	isPodReady := false

	timeAfter := time.After(120 * time.Second)
	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			switch wEvent.Type {
			case kubernetes.EventAdded:
				log.Info().Str("pod", kubernetes.HubPodName).Msg("Added:")
			case kubernetes.EventDeleted:
				log.Info().Str("pod", kubernetes.HubPodName).Msg("Removed:")
				cancel()
				return
			case kubernetes.EventModified:
				modifiedPod, err := wEvent.ToPod()
				if err != nil {
					log.Error().Str("pod", kubernetes.HubPodName).Err(err).Msg("While watching pod.")
					cancel()
					continue
				}

				log.Debug().
					Str("pod", kubernetes.HubPodName).
					Interface("phase", modifiedPod.Status.Phase).
					Interface("containers-statuses", modifiedPod.Status.ContainerStatuses).
					Msg("Watching pod.")

				if modifiedPod.Status.Phase == core.PodRunning && !isPodReady {
					isPodReady = true

					ready.Lock()
					ready.Hub = true
					ready.Unlock()
					postHubStarted(ctx, kubernetesProvider, cancel, false)
				}

				ready.Lock()
				proxyDone := ready.Proxy
				hubPodReady := ready.Hub
				frontPodReady := ready.Front
				ready.Unlock()

				if !proxyDone && hubPodReady && frontPodReady {
					ready.Lock()
					ready.Proxy = true
					ready.Unlock()
					postFrontStarted(ctx, kubernetesProvider, cancel)
				}
			case kubernetes.EventBookmark:
				break
			case kubernetes.EventError:
				break
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			log.Error().
				Str("pod", kubernetes.HubPodName).
				Str("namespace", config.Config.Tap.SelfNamespace).
				Err(err).
				Msg("Failed creating pod.")
			cancel()

		case <-timeAfter:
			if !isPodReady {
				log.Error().
					Str("pod", kubernetes.HubPodName).
					Msg("Pod was not ready in time.")
				cancel()
			}
		case <-ctx.Done():
			log.Debug().
				Str("pod", kubernetes.HubPodName).
				Msg("Watching pod, context done.")
			return
		}
	}
}

func watchFrontPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", kubernetes.FrontPodName))
	podWatchHelper := kubernetes.NewPodWatchHelper(kubernetesProvider, podExactRegex)
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, podWatchHelper, []string{config.Config.Tap.SelfNamespace}, podWatchHelper)
	isPodReady := false

	timeAfter := time.After(120 * time.Second)
	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			switch wEvent.Type {
			case kubernetes.EventAdded:
				log.Info().Str("pod", kubernetes.FrontPodName).Msg("Added:")
			case kubernetes.EventDeleted:
				log.Info().Str("pod", kubernetes.FrontPodName).Msg("Removed:")
				cancel()
				return
			case kubernetes.EventModified:
				modifiedPod, err := wEvent.ToPod()
				if err != nil {
					log.Error().Str("pod", kubernetes.FrontPodName).Err(err).Msg("While watching pod.")
					cancel()
					continue
				}

				log.Debug().
					Str("pod", kubernetes.FrontPodName).
					Interface("phase", modifiedPod.Status.Phase).
					Interface("containers-statuses", modifiedPod.Status.ContainerStatuses).
					Msg("Watching pod.")

				if modifiedPod.Status.Phase == core.PodRunning && !isPodReady {
					isPodReady = true
					ready.Lock()
					ready.Front = true
					ready.Unlock()
				}

				ready.Lock()
				proxyDone := ready.Proxy
				hubPodReady := ready.Hub
				frontPodReady := ready.Front
				ready.Unlock()

				if !proxyDone && hubPodReady && frontPodReady {
					ready.Lock()
					ready.Proxy = true
					ready.Unlock()
					postFrontStarted(ctx, kubernetesProvider, cancel)
				}
			case kubernetes.EventBookmark:
				break
			case kubernetes.EventError:
				break
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			log.Error().
				Str("pod", kubernetes.FrontPodName).
				Str("namespace", config.Config.Tap.SelfNamespace).
				Err(err).
				Msg("Failed creating pod.")
			cancel()

		case <-timeAfter:
			if !isPodReady {
				log.Error().
					Str("pod", kubernetes.FrontPodName).
					Msg("Pod was not ready in time.")
				cancel()
			}
		case <-ctx.Done():
			log.Debug().
				Str("pod", kubernetes.FrontPodName).
				Msg("Watching pod, context done.")
			return
		}
	}
}

func watchHubEvents(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s", kubernetes.HubPodName))
	eventWatchHelper := kubernetes.NewEventWatchHelper(kubernetesProvider, podExactRegex, "pod")
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, eventWatchHelper, []string{config.Config.Tap.SelfNamespace}, eventWatchHelper)
	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			event, err := wEvent.ToEvent()
			if err != nil {
				log.Error().
					Str("pod", kubernetes.HubPodName).
					Err(err).
					Msg("Parsing resource event.")
				continue
			}

			if state.startTime.After(event.CreationTimestamp.Time) {
				continue
			}

			log.Debug().
				Str("pod", kubernetes.HubPodName).
				Str("event", event.Name).
				Time("time", event.CreationTimestamp.Time).
				Str("name", event.Regarding.Name).
				Str("kind", event.Regarding.Kind).
				Str("reason", event.Reason).
				Str("note", event.Note).
				Msg("Watching events.")

			switch event.Reason {
			case "FailedScheduling", "Failed":
				log.Error().
					Str("pod", kubernetes.HubPodName).
					Str("event", event.Name).
					Time("time", event.CreationTimestamp.Time).
					Str("name", event.Regarding.Name).
					Str("kind", event.Regarding.Kind).
					Str("reason", event.Reason).
					Str("note", event.Note).
					Msg("Watching events.")
				cancel()

			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			log.Error().
				Str("pod", kubernetes.HubPodName).
				Err(err).
				Msg("While watching events.")

		case <-ctx.Done():
			log.Debug().
				Str("pod", kubernetes.HubPodName).
				Msg("Watching pod events, context done.")
			return
		}
	}
}

func postHubStarted(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc, update bool) {
	startProxyReportErrorIfAny(
		kubernetesProvider,
		ctx,
		kubernetes.HubServiceName,
		kubernetes.HubPodName,
		configStructs.ProxyHubPortLabel,
		config.Config.Tap.Proxy.Hub.SrcPort,
		config.Config.Tap.Proxy.Hub.DstPort,
		"/echo",
	)

	if !update {
		// Create workers
		err := kubernetes.CreateWorkers(
			kubernetesProvider,
			state.selfServiceAccountExists,
			ctx,
			config.Config.Tap.SelfNamespace,
			config.Config.Tap.Resources.Worker,
			config.Config.ImagePullPolicy(),
			config.Config.ImagePullSecrets(),
			config.Config.Tap.ServiceMesh,
			config.Config.Tap.Tls,
			config.Config.Tap.Debug,
		)
		if err != nil {
			log.Error().Err(err).Send()
		}
	} else {
		// Pod regex
		connector.PostRegexToHub(config.Config.Tap.PodRegexStr, state.targetNamespaces)

		// License
		if config.Config.License != "" {
			connector.PostLicense(config.Config.License)
		}

		// Scripting
		connector.PostEnv(config.Config.Scripting.Env)

		scripts, err := config.Config.Scripting.GetScripts()
		if err != nil {
			log.Error().Err(err).Send()
		}

		for _, script := range scripts {
			_, err = connector.PostScript(script)
			if err != nil {
				log.Error().Err(err).Send()
			}
		}

		connector.PostScriptDone()
	}

	if !update {
		// Hub proxy URL
		url := kubernetes.GetLocalhostOnPort(config.Config.Tap.Proxy.Hub.SrcPort)
		log.Info().Str("url", url).Msg(fmt.Sprintf(utils.Green, "Hub is available at:"))
	}

	if config.Config.Scripting.Source != "" && config.Config.Scripting.WatchScripts {
		watchScripts(false)
	}
}

func postFrontStarted(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	startProxyReportErrorIfAny(
		kubernetesProvider,
		ctx,
		kubernetes.FrontServiceName,
		kubernetes.FrontPodName,
		configStructs.ProxyFrontPortLabel,
		config.Config.Tap.Proxy.Front.SrcPort,
		config.Config.Tap.Proxy.Front.DstPort,
		"",
	)

	url := kubernetes.GetLocalhostOnPort(config.Config.Tap.Proxy.Front.SrcPort)
	log.Info().Str("url", url).Msg(fmt.Sprintf(utils.Green, fmt.Sprintf("%s is available at:", misc.Software)))

	if !config.Config.HeadlessMode {
		utils.OpenBrowser(url)
	}
}
