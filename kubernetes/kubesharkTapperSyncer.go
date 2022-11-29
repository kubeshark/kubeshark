package kubernetes

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/kubeshark/kubeshark/debounce"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/kubeshark/worker/api"
	"github.com/kubeshark/worker/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
)

const updateTappersDelay = 5 * time.Second

type TappedPodChangeEvent struct {
	Added   []core.Pod
	Removed []core.Pod
}

// KubesharkTapperSyncer uses a k8s pod watch to update tapper daemonsets when targeted pods are removed or created
type KubesharkTapperSyncer struct {
	startTime              time.Time
	context                context.Context
	CurrentlyTappedPods    []core.Pod
	config                 TapperSyncerConfig
	kubernetesProvider     *Provider
	TapPodChangesOut       chan TappedPodChangeEvent
	TapperStatusChangedOut chan models.TapperStatus
	ErrorOut               chan K8sTapManagerError
	nodeToTappedPodMap     models.NodeToPodsMap
	tappedNodes            []string
}

type TapperSyncerConfig struct {
	TargetNamespaces              []string
	PodFilterRegex                regexp.Regexp
	KubesharkResourcesNamespace   string
	TapperResources               models.Resources
	ImagePullPolicy               core.PullPolicy
	LogLevel                      zerolog.Level
	KubesharkApiFilteringOptions  api.TrafficFilteringOptions
	KubesharkServiceAccountExists bool
	ServiceMesh                   bool
	Tls                           bool
	MaxLiveStreams                int
}

func CreateAndStartKubesharkTapperSyncer(ctx context.Context, kubernetesProvider *Provider, config TapperSyncerConfig, startTime time.Time) (*KubesharkTapperSyncer, error) {
	syncer := &KubesharkTapperSyncer{
		startTime:              startTime.Truncate(time.Second), // Round down because k8s CreationTimestamp is given in 1 sec resolution.
		context:                ctx,
		CurrentlyTappedPods:    make([]core.Pod, 0),
		config:                 config,
		kubernetesProvider:     kubernetesProvider,
		TapPodChangesOut:       make(chan TappedPodChangeEvent, 100),
		TapperStatusChangedOut: make(chan models.TapperStatus, 100),
		ErrorOut:               make(chan K8sTapManagerError, 100),
	}

	if err, _ := syncer.updateCurrentlyTappedPods(); err != nil {
		return nil, err
	}

	if err := syncer.updateKubesharkTappers(); err != nil {
		return nil, err
	}

	go syncer.watchPodsForTapping()
	go syncer.watchTapperEvents()
	go syncer.watchTapperPods()
	return syncer, nil
}

func (tapperSyncer *KubesharkTapperSyncer) watchTapperPods() {
	kubesharkResourceRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", TapperPodName))
	podWatchHelper := NewPodWatchHelper(tapperSyncer.kubernetesProvider, kubesharkResourceRegex)
	eventChan, errorChan := FilteredWatch(tapperSyncer.context, podWatchHelper, []string{tapperSyncer.config.KubesharkResourcesNamespace}, podWatchHelper)

	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			pod, err := wEvent.ToPod()
			if err != nil {
				log.Error().Str("pod", TapperPodName).Err(err).Msg("While parsing Kubeshark resource!")
				continue
			}

			log.Debug().
				Str("pod", pod.Name).
				Str("node", pod.Spec.NodeName).
				Interface("phase", pod.Status.Phase).
				Msg("Watching pod events...")
			if pod.Spec.NodeName != "" {
				tapperStatus := models.TapperStatus{TapperName: pod.Name, NodeName: pod.Spec.NodeName, Status: string(pod.Status.Phase)}
				tapperSyncer.TapperStatusChangedOut <- tapperStatus
			}

		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}
			log.Error().Str("pod", TapperPodName).Err(err).Msg("While watching pod!")

		case <-tapperSyncer.context.Done():
			log.Debug().
				Str("pod", TapperPodName).
				Msg("Watching pod, context done.")
			return
		}
	}
}

func (tapperSyncer *KubesharkTapperSyncer) watchTapperEvents() {
	kubesharkResourceRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", TapperPodName))
	eventWatchHelper := NewEventWatchHelper(tapperSyncer.kubernetesProvider, kubesharkResourceRegex, "pod")
	eventChan, errorChan := FilteredWatch(tapperSyncer.context, eventWatchHelper, []string{tapperSyncer.config.KubesharkResourcesNamespace}, eventWatchHelper)

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
					Str("pod", TapperPodName).
					Err(err).
					Msg("Parsing resource event.")
				continue
			}

			log.Debug().
				Str("pod", TapperPodName).
				Str("event", event.Name).
				Time("time", event.CreationTimestamp.Time).
				Str("name", event.Regarding.Name).
				Str("kind", event.Regarding.Kind).
				Str("reason", event.Reason).
				Str("note", event.Note).
				Msg("Watching events.")

			pod, err1 := tapperSyncer.kubernetesProvider.GetPod(tapperSyncer.context, tapperSyncer.config.KubesharkResourcesNamespace, event.Regarding.Name)
			if err1 != nil {
				log.Error().Str("name", event.Regarding.Name).Msg("Couldn't get pod")
				continue
			}

			nodeName := ""
			if event.Reason != "FailedScheduling" {
				nodeName = pod.Spec.NodeName
			} else {
				nodeName = pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchFields[0].Values[0]
			}

			tapperStatus := models.TapperStatus{TapperName: pod.Name, NodeName: nodeName, Status: string(pod.Status.Phase)}
			tapperSyncer.TapperStatusChangedOut <- tapperStatus

		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			log.Error().
				Str("pod", TapperPodName).
				Err(err).
				Msg("While watching events.")

		case <-tapperSyncer.context.Done():
			log.Debug().
				Str("pod", TapperPodName).
				Msg("Watching pod events, context done.")
			return
		}
	}
}

func (tapperSyncer *KubesharkTapperSyncer) watchPodsForTapping() {
	podWatchHelper := NewPodWatchHelper(tapperSyncer.kubernetesProvider, &tapperSyncer.config.PodFilterRegex)
	eventChan, errorChan := FilteredWatch(tapperSyncer.context, podWatchHelper, tapperSyncer.config.TargetNamespaces, podWatchHelper)

	handleChangeInPods := func() {
		err, changeFound := tapperSyncer.updateCurrentlyTappedPods()
		if err != nil {
			tapperSyncer.ErrorOut <- K8sTapManagerError{
				OriginalError:    err,
				TapManagerReason: TapManagerPodListError,
			}
		}

		if !changeFound {
			log.Debug().Msg("Nothing changed. Updating tappers is not needed.")
			return
		}
		if err := tapperSyncer.updateKubesharkTappers(); err != nil {
			tapperSyncer.ErrorOut <- K8sTapManagerError{
				OriginalError:    err,
				TapManagerReason: TapManagerTapperUpdateError,
			}
		}
	}
	restartTappersDebouncer := debounce.NewDebouncer(updateTappersDelay, handleChangeInPods)

	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			pod, err := wEvent.ToPod()
			if err != nil {
				tapperSyncer.handleErrorInWatchLoop(err, restartTappersDebouncer)
				continue
			}

			switch wEvent.Type {
			case EventAdded:
				log.Debug().
					Str("pod", pod.Name).
					Str("namespace", pod.Namespace).
					Msg("Added matching pod.")
				if err := restartTappersDebouncer.SetOn(); err != nil {
					log.Error().
						Str("pod", pod.Name).
						Str("namespace", pod.Namespace).
						Err(err).
						Msg("While restarting tappers!")
				}
			case EventDeleted:
				log.Debug().
					Str("pod", pod.Name).
					Str("namespace", pod.Namespace).
					Msg("Removed matching pod.")
				if err := restartTappersDebouncer.SetOn(); err != nil {
					log.Error().
						Str("pod", pod.Name).
						Str("namespace", pod.Namespace).
						Err(err).
						Msg("While restarting tappers!")
				}
			case EventModified:
				log.Debug().
					Str("pod", pod.Name).
					Str("namespace", pod.Namespace).
					Str("ip", pod.Status.PodIP).
					Interface("phase", pod.Status.Phase).
					Msg("Modified matching pod.")

				// Act only if the modified pod has already obtained an IP address.
				// After filtering for IPs, on a normal pod restart this includes the following events:
				// - Pod deletion
				// - Pod reaches start state
				// - Pod reaches ready state
				// Ready/unready transitions might also trigger this event.
				if pod.Status.PodIP != "" {
					if err := restartTappersDebouncer.SetOn(); err != nil {
						log.Error().
							Str("pod", pod.Name).
							Str("namespace", pod.Namespace).
							Err(err).
							Msg("While restarting tappers!")
					}
				}
			case EventBookmark:
				break
			case EventError:
				break
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			tapperSyncer.handleErrorInWatchLoop(err, restartTappersDebouncer)
			continue

		case <-tapperSyncer.context.Done():
			log.Debug().Msg("Watching pods, context done. Stopping \"restart tappers debouncer\"")
			restartTappersDebouncer.Cancel()
			// TODO: Does this also perform cleanup?
			return
		}
	}
}

func (tapperSyncer *KubesharkTapperSyncer) handleErrorInWatchLoop(err error, restartTappersDebouncer *debounce.Debouncer) {
	log.Error().Err(err).Msg("While watching pods, got an error! Stopping \"restart tappers debouncer\"")
	restartTappersDebouncer.Cancel()
	tapperSyncer.ErrorOut <- K8sTapManagerError{
		OriginalError:    err,
		TapManagerReason: TapManagerPodWatchError,
	}
}

func (tapperSyncer *KubesharkTapperSyncer) updateCurrentlyTappedPods() (err error, changesFound bool) {
	if matchingPods, err := tapperSyncer.kubernetesProvider.ListAllRunningPodsMatchingRegex(tapperSyncer.context, &tapperSyncer.config.PodFilterRegex, tapperSyncer.config.TargetNamespaces); err != nil {
		return err, false
	} else {
		podsToTap := excludeKubesharkPods(matchingPods)
		addedPods, removedPods := getPodArrayDiff(tapperSyncer.CurrentlyTappedPods, podsToTap)
		for _, addedPod := range addedPods {
			log.Info().Str("pod", addedPod.Name).Msg("Tapping new pod.")
		}
		for _, removedPod := range removedPods {
			log.Info().Str("pod", removedPod.Name).Msg("Pod is no longer running. Tapping is stopped.")
		}
		if len(addedPods) > 0 || len(removedPods) > 0 {
			tapperSyncer.CurrentlyTappedPods = podsToTap
			tapperSyncer.nodeToTappedPodMap = GetNodeHostToTappedPodsMap(tapperSyncer.CurrentlyTappedPods)
			tapperSyncer.TapPodChangesOut <- TappedPodChangeEvent{
				Added:   addedPods,
				Removed: removedPods,
			}
			return nil, true
		}
		return nil, false
	}
}

func (tapperSyncer *KubesharkTapperSyncer) updateKubesharkTappers() error {
	nodesToTap := make([]string, len(tapperSyncer.nodeToTappedPodMap))
	i := 0
	for node := range tapperSyncer.nodeToTappedPodMap {
		nodesToTap[i] = node
		i++
	}

	if utils.EqualStringSlices(nodesToTap, tapperSyncer.tappedNodes) {
		log.Debug().Msg("Skipping apply, DaemonSet is up to date")
		return nil
	}

	log.Debug().Strs("nodes", nodesToTap).Msg("Updating DaemonSet to run on nodes.")

	image := "kubeshark/worker:latest"

	if len(tapperSyncer.nodeToTappedPodMap) > 0 {
		var serviceAccountName string
		if tapperSyncer.config.KubesharkServiceAccountExists {
			serviceAccountName = ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		nodeNames := make([]string, 0, len(tapperSyncer.nodeToTappedPodMap))
		for nodeName := range tapperSyncer.nodeToTappedPodMap {
			nodeNames = append(nodeNames, nodeName)
		}

		if err := tapperSyncer.kubernetesProvider.ApplyKubesharkTapperDaemonSet(
			tapperSyncer.context,
			tapperSyncer.config.KubesharkResourcesNamespace,
			TapperDaemonSetName,
			image,
			TapperPodName,
			fmt.Sprintf("%s.%s.svc", HubPodName, tapperSyncer.config.KubesharkResourcesNamespace),
			nodeNames,
			serviceAccountName,
			tapperSyncer.config.TapperResources,
			tapperSyncer.config.ImagePullPolicy,
			tapperSyncer.config.KubesharkApiFilteringOptions,
			tapperSyncer.config.LogLevel,
			tapperSyncer.config.ServiceMesh,
			tapperSyncer.config.Tls,
			tapperSyncer.config.MaxLiveStreams); err != nil {
			return err
		}

		log.Debug().Int("tapper-count", len(tapperSyncer.nodeToTappedPodMap)).Msg("Successfully created tappers.")
	} else {
		if err := tapperSyncer.kubernetesProvider.ResetKubesharkTapperDaemonSet(
			tapperSyncer.context,
			tapperSyncer.config.KubesharkResourcesNamespace,
			TapperDaemonSetName,
			image,
			TapperPodName); err != nil {
			return err
		}

		log.Debug().Msg("Successfully reset tapper daemon set")
	}

	tapperSyncer.tappedNodes = nodesToTap

	return nil
}
