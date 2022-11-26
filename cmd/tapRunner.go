package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/resources"
	"github.com/kubeshark/kubeshark/utils"

	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeshark/kubeshark/cmd/goUtils"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/worker/api"
	"github.com/kubeshark/worker/models"
)

const cleanupTimeout = time.Minute

type tapState struct {
	startTime                     time.Time
	targetNamespaces              []string
	kubesharkServiceAccountExists bool
}

var state tapState
var connector *connect.Connector
var hubPodReady bool
var frontPodReady bool
var proxyDone bool

func RunKubesharkTap() {
	state.startTime = time.Now()

	connector = connect.NewConnector(kubernetes.GetLocalhostOnPort(config.Config.Hub.PortForward.SrcPort), connect.DefaultRetries, connect.DefaultTimeout)

	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	state.targetNamespaces = getNamespaces(kubernetesProvider)

	kubesharkAgentConfig := getTapKubesharkAgentConfig()
	serializedKubesharkConfig, err := getSerializedKubesharkAgentConfig(kubesharkAgentConfig)
	if err != nil {
		log.Printf(utils.Error, fmt.Sprintf("Error serializing kubeshark config: %v", errormessage.FormatError(err)))
		return
	}

	if config.Config.IsNsRestrictedMode() {
		if len(state.targetNamespaces) != 1 || !utils.Contains(state.targetNamespaces, config.Config.KubesharkResourcesNamespace) {
			log.Printf("Not supported mode. Kubeshark can't resolve IPs in other namespaces when running in namespace restricted mode.\n"+
				"You can use the same namespace for --%s and --%s", configStructs.NamespacesTapName, config.KubesharkResourcesNamespaceConfigName)
			return
		}
	}

	var namespacesStr string
	if !utils.Contains(state.targetNamespaces, kubernetes.K8sAllNamespaces) {
		namespacesStr = fmt.Sprintf("namespaces \"%s\"", strings.Join(state.targetNamespaces, "\", \""))
	} else {
		namespacesStr = "all namespaces"
	}

	log.Printf("Tapping pods in %s", namespacesStr)

	if err := printTappedPodsPreview(ctx, kubernetesProvider, state.targetNamespaces); err != nil {
		log.Printf(utils.Error, fmt.Sprintf("Error listing pods: %v", errormessage.FormatError(err)))
	}

	if config.Config.Tap.DryRun {
		return
	}

	log.Printf("Waiting for Kubeshark Agent to start...")
	if state.kubesharkServiceAccountExists, err = resources.CreateTapKubesharkResources(ctx, kubernetesProvider, serializedKubesharkConfig, config.Config.IsNsRestrictedMode(), config.Config.KubesharkResourcesNamespace, config.Config.AgentImage, config.Config.Tap.MaxEntriesDBSizeBytes(), config.Config.Tap.HubResources, config.Config.ImagePullPolicy(), config.Config.LogLevel(), config.Config.Tap.Profiler); err != nil {
		var statusError *k8serrors.StatusError
		if errors.As(err, &statusError) && (statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists) {
			log.Print("Kubeshark is already running in this namespace, change the `kubeshark-resources-namespace` configuration or run `kubeshark clean` to remove the currently running Kubeshark instance")
		} else {
			defer resources.CleanUpKubesharkResources(ctx, cancel, kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.KubesharkResourcesNamespace)
			log.Printf(utils.Error, fmt.Sprintf("Error creating resources: %v", errormessage.FormatError(err)))
		}

		return
	}

	defer finishTapExecution(kubernetesProvider)

	go goUtils.HandleExcWrapper(watchHubEvents, ctx, kubernetesProvider, cancel)
	go goUtils.HandleExcWrapper(watchHubPod, ctx, kubernetesProvider, cancel)
	go goUtils.HandleExcWrapper(watchFrontPod, ctx, kubernetesProvider, cancel)

	// block until exit signal or error
	utils.WaitForFinish(ctx, cancel)
}

func finishTapExecution(kubernetesProvider *kubernetes.Provider) {
	finishKubesharkExecution(kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.KubesharkResourcesNamespace)
}

func getTapKubesharkAgentConfig() *models.Config {
	kubesharkAgentConfig := models.Config{
		MaxDBSizeBytes:              config.Config.Tap.MaxEntriesDBSizeBytes(),
		InsertionFilter:             config.Config.Tap.GetInsertionFilter(),
		AgentImage:                  config.Config.AgentImage,
		PullPolicy:                  config.Config.ImagePullPolicyStr,
		LogLevel:                    config.Config.LogLevel(),
		TapperResources:             config.Config.Tap.TapperResources,
		KubesharkResourcesNamespace: config.Config.KubesharkResourcesNamespace,
		AgentDatabasePath:           models.DataDirPath,
		ServiceMap:                  config.Config.ServiceMap,
		OAS:                         config.Config.OAS,
	}

	return &kubesharkAgentConfig
}

/*
this function is a bit problematic as it might be detached from the actual pods the Kubeshark Hub will tap.
The alternative would be to wait for Hub to be ready and then query it for the pods it listens to, this has
the arguably worse drawback of taking a relatively very long time before the user sees which pods are targeted, if any.
*/
func printTappedPodsPreview(ctx context.Context, kubernetesProvider *kubernetes.Provider, namespaces []string) error {
	if matchingPods, err := kubernetesProvider.ListAllRunningPodsMatchingRegex(ctx, config.Config.Tap.PodRegex(), namespaces); err != nil {
		return err
	} else {
		if len(matchingPods) == 0 {
			printNoPodsFoundSuggestion(namespaces)
		}
		for _, tappedPod := range matchingPods {
			log.Printf(utils.Green, fmt.Sprintf("+%s", tappedPod.Name))
		}
		return nil
	}
}

func startTapperSyncer(ctx context.Context, cancel context.CancelFunc, provider *kubernetes.Provider, targetNamespaces []string, startTime time.Time) error {
	tapperSyncer, err := kubernetes.CreateAndStartKubesharkTapperSyncer(ctx, provider, kubernetes.TapperSyncerConfig{
		TargetNamespaces:            targetNamespaces,
		PodFilterRegex:              *config.Config.Tap.PodRegex(),
		KubesharkResourcesNamespace: config.Config.KubesharkResourcesNamespace,
		AgentImage:                  config.Config.AgentImage,
		TapperResources:             config.Config.Tap.TapperResources,
		ImagePullPolicy:             config.Config.ImagePullPolicy(),
		LogLevel:                    config.Config.LogLevel(),
		KubesharkApiFilteringOptions: api.TrafficFilteringOptions{
			IgnoredUserAgents: config.Config.Tap.IgnoredUserAgents,
		},
		KubesharkServiceAccountExists: state.kubesharkServiceAccountExists,
		ServiceMesh:                   config.Config.Tap.ServiceMesh,
		Tls:                           config.Config.Tap.Tls,
		MaxLiveStreams:                config.Config.Tap.MaxLiveStreams,
	}, startTime)

	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case syncerErr, ok := <-tapperSyncer.ErrorOut:
				if !ok {
					log.Print("kubesharkTapperSyncer err channel closed, ending listener loop")
					return
				}
				log.Printf(utils.Error, getErrorDisplayTextForK8sTapManagerError(syncerErr))
				cancel()
			case _, ok := <-tapperSyncer.TapPodChangesOut:
				if !ok {
					log.Print("kubesharkTapperSyncer pod changes channel closed, ending listener loop")
					return
				}
				if err := connector.ReportTappedPods(tapperSyncer.CurrentlyTappedPods); err != nil {
					log.Printf("[Error] failed update tapped pods %v", err)
				}
			case tapperStatus, ok := <-tapperSyncer.TapperStatusChangedOut:
				if !ok {
					log.Print("kubesharkTapperSyncer tapper status changed channel closed, ending listener loop")
					return
				}
				if err := connector.ReportTapperStatus(tapperStatus); err != nil {
					log.Printf("[Error] failed update tapper status %v", err)
				}
			case <-ctx.Done():
				log.Print("kubesharkTapperSyncer event listener loop exiting due to context done")
				return
			}
		}
	}()

	return nil
}

func printNoPodsFoundSuggestion(targetNamespaces []string) {
	var suggestionStr string
	if !utils.Contains(targetNamespaces, kubernetes.K8sAllNamespaces) {
		suggestionStr = ". You can also try selecting a different namespace with -n or tap all namespaces with -A"
	}
	log.Printf(utils.Warning, fmt.Sprintf("Did not find any currently running pods that match the regex argument, kubeshark will automatically tap matching pods if any are created later%s", suggestionStr))
}

func getErrorDisplayTextForK8sTapManagerError(err kubernetes.K8sTapManagerError) string {
	switch err.TapManagerReason {
	case kubernetes.TapManagerPodListError:
		return fmt.Sprintf("Failed to update currently tapped pods: %v", err.OriginalError)
	case kubernetes.TapManagerPodWatchError:
		return fmt.Sprintf("Error occured in k8s pod watch: %v", err.OriginalError)
	case kubernetes.TapManagerTapperUpdateError:
		return fmt.Sprintf("Error updating tappers: %v", err.OriginalError)
	default:
		return fmt.Sprintf("Unknown error occured in k8s tap manager: %v", err.OriginalError)
	}
}

func watchHubPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", kubernetes.HubPodName))
	podWatchHelper := kubernetes.NewPodWatchHelper(kubernetesProvider, podExactRegex)
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, podWatchHelper, []string{config.Config.KubesharkResourcesNamespace}, podWatchHelper)
	isPodReady := false

	hubTimeoutSec := config.GetIntEnvConfig(config.HubTimeoutSec, 120)
	timeAfter := time.After(time.Duration(hubTimeoutSec) * time.Second)
	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			switch wEvent.Type {
			case kubernetes.EventAdded:
				log.Printf("Watching Hub pod loop, added")
			case kubernetes.EventDeleted:
				log.Printf("%s removed", kubernetes.HubPodName)
				cancel()
				return
			case kubernetes.EventModified:
				modifiedPod, err := wEvent.ToPod()
				if err != nil {
					log.Printf(utils.Error, err)
					cancel()
					continue
				}

				log.Printf("Watching Hub pod loop, modified: %v, containers statuses: %v", modifiedPod.Status.Phase, modifiedPod.Status.ContainerStatuses)

				if modifiedPod.Status.Phase == core.PodRunning && !isPodReady {
					isPodReady = true
					hubPodReady = true
					postHubStarted(ctx, kubernetesProvider, cancel)
				}

				if !proxyDone && hubPodReady && frontPodReady {
					proxyDone = true
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

			log.Printf("[ERROR] Agent creation, watching %v namespace, error: %v", config.Config.KubesharkResourcesNamespace, err)
			cancel()

		case <-timeAfter:
			if !isPodReady {
				log.Printf(utils.Error, "Kubeshark Hub was not ready in time")
				cancel()
			}
		case <-ctx.Done():
			log.Printf("Watching Hub pod loop, ctx done")
			return
		}
	}
}

func watchFrontPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s$", kubernetes.FrontPodName))
	podWatchHelper := kubernetes.NewPodWatchHelper(kubernetesProvider, podExactRegex)
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, podWatchHelper, []string{config.Config.KubesharkResourcesNamespace}, podWatchHelper)
	isPodReady := false

	hubTimeoutSec := config.GetIntEnvConfig(config.HubTimeoutSec, 120)
	timeAfter := time.After(time.Duration(hubTimeoutSec) * time.Second)
	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			switch wEvent.Type {
			case kubernetes.EventAdded:
				log.Printf("Watching Hub pod loop, added")
			case kubernetes.EventDeleted:
				log.Printf("%s removed", kubernetes.FrontPodName)
				cancel()
				return
			case kubernetes.EventModified:
				modifiedPod, err := wEvent.ToPod()
				if err != nil {
					log.Printf(utils.Error, err)
					cancel()
					continue
				}

				log.Printf("Watching Hub pod loop, modified: %v, containers statuses: %v", modifiedPod.Status.Phase, modifiedPod.Status.ContainerStatuses)

				if modifiedPod.Status.Phase == core.PodRunning && !isPodReady {
					isPodReady = true
					frontPodReady = true
				}

				if !proxyDone && hubPodReady && frontPodReady {
					proxyDone = true
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

			log.Printf("[ERROR] Agent creation, watching %v namespace, error: %v", config.Config.KubesharkResourcesNamespace, err)
			cancel()

		case <-timeAfter:
			if !isPodReady {
				log.Printf(utils.Error, "Kubeshark Hub was not ready in time")
				cancel()
			}
		case <-ctx.Done():
			log.Printf("Watching Hub pod loop, ctx done")
			return
		}
	}
}

func watchHubEvents(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s", kubernetes.HubPodName))
	eventWatchHelper := kubernetes.NewEventWatchHelper(kubernetesProvider, podExactRegex, "pod")
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, eventWatchHelper, []string{config.Config.KubesharkResourcesNamespace}, eventWatchHelper)
	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			event, err := wEvent.ToEvent()
			if err != nil {
				log.Printf("[ERROR] parsing Kubeshark resource event: %+v", err)
				continue
			}

			if state.startTime.After(event.CreationTimestamp.Time) {
				continue
			}

			log.Printf(
				"Watching Hub events loop, event %s, time: %v, resource: %s (%s), reason: %s, note: %s",
				event.Name,
				event.CreationTimestamp.Time,
				event.Regarding.Name,
				event.Regarding.Kind,
				event.Reason,
				event.Note,
			)

			switch event.Reason {
			case "FailedScheduling", "Failed":
				log.Printf(utils.Error, fmt.Sprintf("Kubeshark Hub status: %s - %s", event.Reason, event.Note))
				cancel()

			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			log.Printf("[Error] Watching Hub events loop, error: %+v", err)
		case <-ctx.Done():
			log.Printf("Watching Hub events loop, ctx done")
			return
		}
	}
}

func postHubStarted(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	startProxyReportErrorIfAny(kubernetesProvider, ctx, cancel, kubernetes.HubServiceName, config.Config.Hub.PortForward.SrcPort, config.Config.Hub.PortForward.DstPort, "/echo")

	if err := startTapperSyncer(ctx, cancel, kubernetesProvider, state.targetNamespaces, state.startTime); err != nil {
		log.Printf(utils.Error, fmt.Sprintf("Error starting kubeshark tapper syncer: %v", errormessage.FormatError(err)))
		cancel()
	}

	url := kubernetes.GetLocalhostOnPort(config.Config.Hub.PortForward.SrcPort)
	log.Printf("Hub is available at %s", url)
}

func postFrontStarted(ctx context.Context, kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	startProxyReportErrorIfAny(kubernetesProvider, ctx, cancel, kubernetes.FrontServiceName, config.Config.Front.PortForward.SrcPort, config.Config.Front.PortForward.DstPort, "")

	url := kubernetes.GetLocalhostOnPort(config.Config.Front.PortForward.SrcPort)
	log.Printf("Kubeshark is available at %s", url)
	if !config.Config.HeadlessMode {
		utils.OpenBrowser(url)
	}
}

func getNamespaces(kubernetesProvider *kubernetes.Provider) []string {
	if config.Config.Tap.AllNamespaces {
		return []string{kubernetes.K8sAllNamespaces}
	} else if len(config.Config.Tap.Namespaces) > 0 {
		return utils.Unique(config.Config.Tap.Namespaces)
	} else {
		currentNamespace, err := kubernetesProvider.CurrentNamespace()
		if err != nil {
			log.Fatalf(utils.Red, fmt.Sprintf("error getting current namespace: %+v", err))
		}
		return []string{currentNamespace}
	}
}
