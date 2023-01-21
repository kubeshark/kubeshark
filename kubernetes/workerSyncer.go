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

type TargetedPodChangeEvent struct {
	Added   []v1.Pod
	Removed []v1.Pod
}

// WorkerSyncer uses a k8s pod watch to update Worker daemonsets when targeted pods are removed or created
type WorkerSyncer struct {
	startTime             time.Time
	context               context.Context
	CurrentlyTargetedPods []v1.Pod
	config                WorkerSyncerConfig
	kubernetesProvider    *Provider
	TapPodChangesOut      chan TargetedPodChangeEvent
	ErrorOut              chan K8sTapManagerError
	nodeToTargetedPodMap  models.NodeToPodsMap
	targetedNodes         []string
}

type WorkerSyncerConfig struct {
	TargetNamespaces         []string
	PodFilterRegex           regexp.Regexp
	SelfNamespace            string
	WorkerResources          Resources
	ImagePullPolicy          v1.PullPolicy
	ImagePullSecrets         []v1.LocalObjectReference
	SelfServiceAccountExists bool
	ServiceMesh              bool
	Tls                      bool
	Debug                    bool
}

func CreateAndStartWorkerSyncer(ctx context.Context, kubernetesProvider *Provider, config WorkerSyncerConfig, startTime time.Time) (*WorkerSyncer, error) {
	syncer := &WorkerSyncer{
		startTime:             startTime.Truncate(time.Second), // Round down because k8s CreationTimestamp is given in 1 sec resolution.
		context:               ctx,
		CurrentlyTargetedPods: make([]v1.Pod, 0),
		config:                config,
		kubernetesProvider:    kubernetesProvider,
		TapPodChangesOut:      make(chan TargetedPodChangeEvent, 100),
		ErrorOut:              make(chan K8sTapManagerError, 100),
	}

	if err, _ := syncer.updateCurrentlyTargetedPods(); err != nil {
		return nil, err
	}

	if err := syncer.updateWorkers(); err != nil {
		return nil, err
	}

	go syncer.watchPodsForTargeting()
	return syncer, nil
}

func (workerSyncer *WorkerSyncer) watchPodsForTargeting() {
	podWatchHelper := NewPodWatchHelper(workerSyncer.kubernetesProvider, &workerSyncer.config.PodFilterRegex)
	eventChan, errorChan := FilteredWatch(workerSyncer.context, podWatchHelper, workerSyncer.config.TargetNamespaces, podWatchHelper)

	handleChangeInPods := func() {
		err, changeFound := workerSyncer.updateCurrentlyTargetedPods()
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

func (workerSyncer *WorkerSyncer) updateCurrentlyTargetedPods() (err error, changesFound bool) {
	if matchingPods, err := workerSyncer.kubernetesProvider.ListAllRunningPodsMatchingRegex(workerSyncer.context, &workerSyncer.config.PodFilterRegex, workerSyncer.config.TargetNamespaces); err != nil {
		return err, false
	} else {
		podsToTarget := excludeSelfPods(matchingPods)
		addedPods, removedPods := getPodArrayDiff(workerSyncer.CurrentlyTargetedPods, podsToTarget)
		for _, addedPod := range addedPods {
			log.Info().Msg(fmt.Sprintf("Targeted pod: %s", fmt.Sprintf(utils.Green, addedPod.Name)))
		}
		for _, removedPod := range removedPods {
			log.Info().Msg(fmt.Sprintf("Untargeted pod: %s", fmt.Sprintf(utils.Red, removedPod.Name)))
		}
		if len(addedPods) > 0 || len(removedPods) > 0 {
			workerSyncer.CurrentlyTargetedPods = podsToTarget
			workerSyncer.nodeToTargetedPodMap = GetNodeHostToTargetedPodsMap(workerSyncer.CurrentlyTargetedPods)
			workerSyncer.TapPodChangesOut <- TargetedPodChangeEvent{
				Added:   addedPods,
				Removed: removedPods,
			}
			return nil, true
		}
		return nil, false
	}
}

func (workerSyncer *WorkerSyncer) updateWorkers() error {
	nodesToTarget := make([]string, len(workerSyncer.nodeToTargetedPodMap))
	i := 0
	for node := range workerSyncer.nodeToTargetedPodMap {
		nodesToTarget[i] = node
		i++
	}

	if utils.EqualStringSlices(nodesToTarget, workerSyncer.targetedNodes) {
		log.Debug().Msg("Skipping apply, DaemonSet is up to date")
		return nil
	}

	log.Debug().Strs("nodes", nodesToTarget).Msg("Updating DaemonSet to run on nodes.")

	image := docker.GetWorkerImage()

	if len(workerSyncer.nodeToTargetedPodMap) > 0 {
		var serviceAccountName string
		if workerSyncer.config.SelfServiceAccountExists {
			serviceAccountName = ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		nodeNames := make([]string, 0, len(workerSyncer.nodeToTargetedPodMap))
		for nodeName := range workerSyncer.nodeToTargetedPodMap {
			nodeNames = append(nodeNames, nodeName)
		}

		if err := workerSyncer.kubernetesProvider.ApplyWorkerDaemonSet(
			workerSyncer.context,
			workerSyncer.config.SelfNamespace,
			WorkerDaemonSetName,
			image,
			WorkerPodName,
			nodeNames,
			serviceAccountName,
			workerSyncer.config.WorkerResources,
			workerSyncer.config.ImagePullPolicy,
			workerSyncer.config.ImagePullSecrets,
			workerSyncer.config.ServiceMesh,
			workerSyncer.config.Tls,
			workerSyncer.config.Debug); err != nil {
			return err
		}

		log.Debug().Int("worker-count", len(workerSyncer.nodeToTargetedPodMap)).Msg("Successfully created workers.")
	} else {
		if err := workerSyncer.kubernetesProvider.ResetWorkerDaemonSet(
			workerSyncer.context,
			workerSyncer.config.SelfNamespace,
			WorkerDaemonSetName,
			image,
			WorkerPodName); err != nil {
			return err
		}

		log.Debug().Msg("Successfully resetted Worker DaemonSet")
	}

	workerSyncer.targetedNodes = nodesToTarget

	return nil
}
