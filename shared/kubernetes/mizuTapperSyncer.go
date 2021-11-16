package kubernetes

import (
	"context"
	"fmt"
	"github.com/op/go-logging"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
	core "k8s.io/api/core/v1"
	"regexp"
	"time"
)

const updateTappersDelay = 5 * time.Second

type TappedPodChangeEvent struct {
	Added   []core.Pod
	Removed []core.Pod
}

// MizuTapperSyncer uses a k8s pod watch to update tapper daemonsets when targeted pods are removed or created
type MizuTapperSyncer struct {
	context             context.Context
	CurrentlyTappedPods []core.Pod
	config              TapperSyncerConfig
	kubernetesProvider  *Provider
	TapPodChangesOut    chan TappedPodChangeEvent
	ErrorOut            chan K8sTapManagerError
}

type TapperSyncerConfig struct {
	TargetNamespaces         []string
	PodFilterRegex           regexp.Regexp
	MizuResourcesNamespace   string
	AgentImage               string
	TapperResources          shared.Resources
	ImagePullPolicy          core.PullPolicy
	LogLevel                 logging.Level
	IgnoredUserAgents        []string
	MizuApiFilteringOptions  api.TrafficFilteringOptions
	MizuServiceAccountExists bool
}

func CreateAndStartMizuTapperSyncer(ctx context.Context, kubernetesProvider *Provider, config TapperSyncerConfig) (*MizuTapperSyncer, error) {
	syncer := &MizuTapperSyncer{
		context:             ctx,
		CurrentlyTappedPods: make([]core.Pod, 0),
		config:              config,
		kubernetesProvider:  kubernetesProvider,
		TapPodChangesOut:    make(chan TappedPodChangeEvent, 100),
		ErrorOut:            make(chan K8sTapManagerError, 100),
	}

	if err, _ := syncer.updateCurrentlyTappedPods(); err != nil {
		return nil, err
	}

	if err := syncer.updateMizuTappers(); err != nil {
		return nil, err
	}

	go syncer.watchPodsForTapping()
	return syncer, nil
}

func (tapperSyncer *MizuTapperSyncer) watchPodsForTapping() {
	podWatchHelper := NewPodWatchHelper(tapperSyncer.kubernetesProvider, &tapperSyncer.config.PodFilterRegex)
	added, modified, removed, errorChan := FilteredWatch(tapperSyncer.context, podWatchHelper, tapperSyncer.config.TargetNamespaces, podWatchHelper)

	restartTappers := func() {
		err, changeFound := tapperSyncer.updateCurrentlyTappedPods()
		if err != nil {
			tapperSyncer.ErrorOut <- K8sTapManagerError{
				OriginalError:    err,
				TapManagerReason: TapManagerPodListError,
			}
		}

		if !changeFound {
			logger.Log.Debugf("Nothing changed update tappers not needed")
			return
		}
		if err := tapperSyncer.updateMizuTappers(); err != nil {
			tapperSyncer.ErrorOut <- K8sTapManagerError{
				OriginalError:    err,
				TapManagerReason: TapManagerTapperUpdateError,
			}
		}
	}
	restartTappersDebouncer := debounce.NewDebouncer(updateTappersDelay, restartTappers)

	for {
		select {
		case event, ok := <-added:
			if !ok {
				added = nil
				continue
			}

			pod, err := podWatchHelper.GetPodFromEvent(event)
			if err != nil {
				tapperSyncer.handleErrorInWatchLoop(err, restartTappersDebouncer)
				continue
			}


			logger.Log.Debugf("Added matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case event, ok := <-removed:
			if !ok {
				removed = nil
				continue
			}

			pod, err := podWatchHelper.GetPodFromEvent(event)
			if err != nil {
				tapperSyncer.handleErrorInWatchLoop(err, restartTappersDebouncer)
				continue
			}

			logger.Log.Debugf("Removed matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case event, ok := <-modified:
			if !ok {
				modified = nil
				continue
			}

			pod, err := podWatchHelper.GetPodFromEvent(event)
			if err != nil {
				tapperSyncer.handleErrorInWatchLoop(err, restartTappersDebouncer)
				continue
			}


			logger.Log.Debugf("Modified matching pod %s, ns: %s, phase: %s, ip: %s", pod.Name, pod.Namespace, pod.Status.Phase, pod.Status.PodIP)
			// Act only if the modified pod has already obtained an IP address.
			// After filtering for IPs, on a normal pod restart this includes the following events:
			// - Pod deletion
			// - Pod reaches start state
			// - Pod reaches ready state
			// Ready/unready transitions might also trigger this event.
			if pod.Status.PodIP != "" {
				restartTappersDebouncer.SetOn()
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			tapperSyncer.handleErrorInWatchLoop(err, restartTappersDebouncer)
			continue

		case <-tapperSyncer.context.Done():
			logger.Log.Debugf("Watching pods loop, context done, stopping `restart tappers debouncer`")
			restartTappersDebouncer.Cancel()
			// TODO: Does this also perform cleanup?
			return
		}
	}
}

func (tapperSyncer *MizuTapperSyncer) handleErrorInWatchLoop(err error, restartTappersDebouncer *debounce.Debouncer) {
	logger.Log.Debugf("Watching pods loop, got error %v, stopping `restart tappers debouncer`", err)
	restartTappersDebouncer.Cancel()
	tapperSyncer.ErrorOut <- K8sTapManagerError{
		OriginalError:    err,
		TapManagerReason: TapManagerPodWatchError,
	}
}

func (tapperSyncer *MizuTapperSyncer) updateCurrentlyTappedPods() (err error, changesFound bool) {
	if matchingPods, err := tapperSyncer.kubernetesProvider.ListAllRunningPodsMatchingRegex(tapperSyncer.context, &tapperSyncer.config.PodFilterRegex, tapperSyncer.config.TargetNamespaces); err != nil {
		return err, false
	} else {
		podsToTap := excludeMizuPods(matchingPods)
		addedPods, removedPods := getPodArrayDiff(tapperSyncer.CurrentlyTappedPods, podsToTap)
		for _, addedPod := range addedPods {
			logger.Log.Debugf("tapping new pod %s", addedPod.Name)
		}
		for _, removedPod := range removedPods {
			logger.Log.Debugf("pod %s is no longer running, tapping for it stopped", removedPod.Name)
		}
		if len(addedPods) > 0 || len(removedPods) > 0 {
			tapperSyncer.CurrentlyTappedPods = podsToTap
			tapperSyncer.TapPodChangesOut <- TappedPodChangeEvent{
				Added:   addedPods,
				Removed: removedPods,
			}
			return nil, true
		}
		return nil, false
	}
}

func (tapperSyncer *MizuTapperSyncer) updateMizuTappers() error {
	nodeToTappedPodIPMap := GetNodeHostToTappedPodIpsMap(tapperSyncer.CurrentlyTappedPods)

	if len(nodeToTappedPodIPMap) > 0 {
		var serviceAccountName string
		if tapperSyncer.config.MizuServiceAccountExists {
			serviceAccountName = ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		if err := tapperSyncer.kubernetesProvider.ApplyMizuTapperDaemonSet(
			tapperSyncer.context,
			tapperSyncer.config.MizuResourcesNamespace,
			TapperDaemonSetName,
			tapperSyncer.config.AgentImage,
			TapperPodName,
			fmt.Sprintf("%s.%s.svc.cluster.local", ApiServerPodName, tapperSyncer.config.MizuResourcesNamespace),
			nodeToTappedPodIPMap,
			serviceAccountName,
			tapperSyncer.config.TapperResources,
			tapperSyncer.config.ImagePullPolicy,
			tapperSyncer.config.MizuApiFilteringOptions,
			tapperSyncer.config.LogLevel,
		); err != nil {
			return err
		}
		logger.Log.Debugf("Successfully created %v tappers", len(nodeToTappedPodIPMap))
	} else {
		if err := tapperSyncer.kubernetesProvider.RemoveDaemonSet(tapperSyncer.context, tapperSyncer.config.MizuResourcesNamespace, TapperDaemonSetName); err != nil {
			return err
		}
	}

	return nil
}
