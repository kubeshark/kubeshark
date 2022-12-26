package kubernetes

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/kubeshark/base/pkg/models"
	"github.com/kubeshark/kubeshark/debounce"
	"github.com/kubeshark/kubeshark/docker"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

const updateWorkersDelay = 5 * time.Second

type TargettedPodChangeEvent struct {
	Added   []v1.Pod
	Removed []v1.Pod
}

// WorkerSyncer uses a k8s pod watch to update Worker daemonsets when targeted pods are removed or created
type WorkerSyncer struct {
	startTime              time.Time
	context                context.Context
	CurrentlyTargettedPods []v1.Pod
	config                 WorkerSyncerConfig
	kubernetesProvider     *Provider
	TapPodChangesOut       chan TargettedPodChangeEvent
	WorkerPodsChanges      chan *v1.Pod
	ErrorOut               chan K8sTapManagerError
	nodeToTargettedPodMap  models.NodeToPodsMap
	targettedNodes         []string
}

type WorkerSyncerConfig struct {
	TargetNamespaces              []string
	PodFilterRegex                regexp.Regexp
	KubesharkResourcesNamespace   string
	WorkerResources               models.Resources
	ImagePullPolicy               v1.PullPolicy
	KubesharkServiceAccountExists bool
	ServiceMesh                   bool
	Tls                           bool
	Debug                         bool
}

func CreateAndStartWorkerSyncer(ctx context.Context, kubernetesProvider *Provider, config WorkerSyncerConfig, startTime time.Time) (*WorkerSyncer, error) {
	syncer := &WorkerSyncer{
		startTime:              startTime.Truncate(time.Second), // Round down because k8s CreationTimestamp is given in 1 sec resolution.
		context:                ctx,
		CurrentlyTargettedPods: make([]v1.Pod, 0),
		config:                 config,
		kubernetesProvider:     kubernetesProvider,
		TapPodChangesOut:       make(chan TargettedPodChangeEvent, 100),
		WorkerPodsChanges:      make(chan *v1.Pod, 100),
		ErrorOut:               make(chan K8sTapManagerError, 100),
	}

	if err, _ := syncer.updateCurrentlyTargettedPods(); err != nil {
		return nil, err
	}

	if err := syncer.updateWorkers(); err != nil {
		return nil, err
	}

	go syncer.watchPodsForTargetting()
	go syncer.watchWorkerEvents()
	go syncer.watchWorkerPods()
	return syncer, nil
}

func (workerSyncer *WorkerSyncer) watchWorkerPods() {
	kubesharkResourceRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", WorkerPodName))
	podWatchHelper := NewPodWatchHelper(workerSyncer.kubernetesProvider, kubesharkResourceRegex)
	eventChan, errorChan := FilteredWatch(workerSyncer.context, podWatchHelper, []string{workerSyncer.config.KubesharkResourcesNamespace}, podWatchHelper)

	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			pod, err := wEvent.ToPod()
			if err != nil {
				log.Error().Str("pod", WorkerPodName).Err(err).Msg("While parsing Kubeshark resource!")
				continue
			}

			log.Debug().
				Str("pod", pod.Name).
				Str("node", pod.Spec.NodeName).
				Interface("phase", pod.Status.Phase).
				Msg("Watching pod events...")
			if pod.Spec.NodeName != "" {
				workerSyncer.WorkerPodsChanges <- pod
			}

		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}
			log.Error().Str("pod", WorkerPodName).Err(err).Msg("While watching pod!")

		case <-workerSyncer.context.Done():
			log.Debug().
				Str("pod", WorkerPodName).
				Msg("Watching pod, context done.")
			return
		}
	}
}

func (workerSyncer *WorkerSyncer) watchWorkerEvents() {
	kubesharkResourceRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", WorkerPodName))
	eventWatchHelper := NewEventWatchHelper(workerSyncer.kubernetesProvider, kubesharkResourceRegex, "pod")
	eventChan, errorChan := FilteredWatch(workerSyncer.context, eventWatchHelper, []string{workerSyncer.config.KubesharkResourcesNamespace}, eventWatchHelper)

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
					Str("pod", WorkerPodName).
					Err(err).
					Msg("Parsing resource event.")
				continue
			}

			log.Debug().
				Str("pod", WorkerPodName).
				Str("event", event.Name).
				Time("time", event.CreationTimestamp.Time).
				Str("name", event.Regarding.Name).
				Str("kind", event.Regarding.Kind).
				Str("reason", event.Reason).
				Str("note", event.Note).
				Msg("Watching events.")

			pod, err1 := workerSyncer.kubernetesProvider.GetPod(workerSyncer.context, workerSyncer.config.KubesharkResourcesNamespace, event.Regarding.Name)
			if err1 != nil {
				log.Error().Str("name", event.Regarding.Name).Msg("Couldn't get pod")
				continue
			}

			workerSyncer.WorkerPodsChanges <- pod

		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			log.Error().
				Str("pod", WorkerPodName).
				Err(err).
				Msg("While watching events.")

		case <-workerSyncer.context.Done():
			log.Debug().
				Str("pod", WorkerPodName).
				Msg("Watching pod events, context done.")
			return
		}
	}
}

func (workerSyncer *WorkerSyncer) watchPodsForTargetting() {
	podWatchHelper := NewPodWatchHelper(workerSyncer.kubernetesProvider, &workerSyncer.config.PodFilterRegex)
	eventChan, errorChan := FilteredWatch(workerSyncer.context, podWatchHelper, workerSyncer.config.TargetNamespaces, podWatchHelper)

	handleChangeInPods := func() {
		err, changeFound := workerSyncer.updateCurrentlyTargettedPods()
		if err != nil {
			workerSyncer.ErrorOut <- K8sTapManagerError{
				OriginalError:    err,
				TapManagerReason: TapManagerPodListError,
			}
		}

		if !changeFound {
			log.Debug().Msg("Nothing changed. Updating workers is not needed.")
			return
		}
		if err := workerSyncer.updateWorkers(); err != nil {
			workerSyncer.ErrorOut <- K8sTapManagerError{
				OriginalError:    err,
				TapManagerReason: TapManagerWorkerUpdateError,
			}
		}
	}
	restartWorkersDebouncer := debounce.NewDebouncer(updateWorkersDelay, handleChangeInPods)

	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			pod, err := wEvent.ToPod()
			if err != nil {
				workerSyncer.handleErrorInWatchLoop(err, restartWorkersDebouncer)
				continue
			}

			switch wEvent.Type {
			case EventAdded:
				log.Debug().
					Str("pod", pod.Name).
					Str("namespace", pod.Namespace).
					Msg("Added matching pod.")
				if err := restartWorkersDebouncer.SetOn(); err != nil {
					log.Error().
						Str("pod", pod.Name).
						Str("namespace", pod.Namespace).
						Err(err).
						Msg("While restarting workers!")
				}
			case EventDeleted:
				log.Debug().
					Str("pod", pod.Name).
					Str("namespace", pod.Namespace).
					Msg("Removed matching pod.")
				if err := restartWorkersDebouncer.SetOn(); err != nil {
					log.Error().
						Str("pod", pod.Name).
						Str("namespace", pod.Namespace).
						Err(err).
						Msg("While restarting workers!")
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
					if err := restartWorkersDebouncer.SetOn(); err != nil {
						log.Error().
							Str("pod", pod.Name).
							Str("namespace", pod.Namespace).
							Err(err).
							Msg("While restarting workers!")
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

			workerSyncer.handleErrorInWatchLoop(err, restartWorkersDebouncer)
			continue

		case <-workerSyncer.context.Done():
			log.Debug().Msg("Watching pods, context done. Stopping \"restart workers debouncer\"")
			restartWorkersDebouncer.Cancel()
			// TODO: Does this also perform cleanup?
			return
		}
	}
}

func (workerSyncer *WorkerSyncer) handleErrorInWatchLoop(err error, restartWorkersDebouncer *debounce.Debouncer) {
	log.Error().Err(err).Msg("While watching pods, got an error! Stopping \"restart workers debouncer\"")
	restartWorkersDebouncer.Cancel()
	workerSyncer.ErrorOut <- K8sTapManagerError{
		OriginalError:    err,
		TapManagerReason: TapManagerPodWatchError,
	}
}

func (workerSyncer *WorkerSyncer) updateCurrentlyTargettedPods() (err error, changesFound bool) {
	if matchingPods, err := workerSyncer.kubernetesProvider.ListAllRunningPodsMatchingRegex(workerSyncer.context, &workerSyncer.config.PodFilterRegex, workerSyncer.config.TargetNamespaces); err != nil {
		return err, false
	} else {
		podsToTarget := excludeSelfPods(matchingPods)
		addedPods, removedPods := getPodArrayDiff(workerSyncer.CurrentlyTargettedPods, podsToTarget)
		for _, addedPod := range addedPods {
			log.Info().Str("pod", addedPod.Name).Msg("Currently targetting:")
		}
		for _, removedPod := range removedPods {
			log.Info().Str("pod", removedPod.Name).Msg("Pod is no longer running. Targetting is stopped.")
		}
		if len(addedPods) > 0 || len(removedPods) > 0 {
			workerSyncer.CurrentlyTargettedPods = podsToTarget
			workerSyncer.nodeToTargettedPodMap = GetNodeHostToTargettedPodsMap(workerSyncer.CurrentlyTargettedPods)
			workerSyncer.TapPodChangesOut <- TargettedPodChangeEvent{
				Added:   addedPods,
				Removed: removedPods,
			}
			return nil, true
		}
		return nil, false
	}
}

func (workerSyncer *WorkerSyncer) updateWorkers() error {
	nodesToTarget := make([]string, len(workerSyncer.nodeToTargettedPodMap))
	i := 0
	for node := range workerSyncer.nodeToTargettedPodMap {
		nodesToTarget[i] = node
		i++
	}

	if utils.EqualStringSlices(nodesToTarget, workerSyncer.targettedNodes) {
		log.Debug().Msg("Skipping apply, DaemonSet is up to date")
		return nil
	}

	log.Debug().Strs("nodes", nodesToTarget).Msg("Updating DaemonSet to run on nodes.")

	image := docker.GetWorkerImage()

	if len(workerSyncer.nodeToTargettedPodMap) > 0 {
		var serviceAccountName string
		if workerSyncer.config.KubesharkServiceAccountExists {
			serviceAccountName = ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		nodeNames := make([]string, 0, len(workerSyncer.nodeToTargettedPodMap))
		for nodeName := range workerSyncer.nodeToTargettedPodMap {
			nodeNames = append(nodeNames, nodeName)
		}

		if err := workerSyncer.kubernetesProvider.ApplyWorkerDaemonSet(
			workerSyncer.context,
			workerSyncer.config.KubesharkResourcesNamespace,
			WorkerDaemonSetName,
			image,
			WorkerPodName,
			nodeNames,
			serviceAccountName,
			workerSyncer.config.WorkerResources,
			workerSyncer.config.ImagePullPolicy,
			workerSyncer.config.ServiceMesh,
			workerSyncer.config.Tls,
			workerSyncer.config.Debug); err != nil {
			return err
		}

		log.Debug().Int("worker-count", len(workerSyncer.nodeToTargettedPodMap)).Msg("Successfully created workers.")
	} else {
		if err := workerSyncer.kubernetesProvider.ResetWorkerDaemonSet(
			workerSyncer.context,
			workerSyncer.config.KubesharkResourcesNamespace,
			WorkerDaemonSetName,
			image,
			WorkerPodName); err != nil {
			return err
		}

		log.Debug().Msg("Successfully resetted Worker DaemonSet")
	}

	workerSyncer.targettedNodes = nodesToTarget

	return nil
}
