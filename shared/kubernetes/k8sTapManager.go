package kubernetes

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	"github.com/up9inc/mizu/shared/goUtils"
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

type K8sTapManager struct {
	context             context.Context
	CurrentlyTappedPods []core.Pod
	config              TapManagerConfig
	kubernetesProvider  *Provider
	TapPodChangesOut    chan TappedPodChangeEvent
	ErrorOut            chan K8sTapManagerError
	shouldUpdateTappers bool // used to prevent updating tapper daemonsets before api is available
}

type TapManagerConfig struct {
	TargetNamespaces         []string
	PodFilterRegex           regexp.Regexp
	MizuResourcesNamespace   string
	AgentImage               string
	TapperResources          shared.Resources
	ImagePullPolicy          core.PullPolicy
	DumpLogs                 bool
	IgnoredUserAgents        []string
	MizuApiFilteringOptions  api.TrafficFilteringOptions
	MizuServiceAccountExists bool
}

func CreateAndStartK8sTapManager(ctx context.Context, kubernetesProvider *Provider, config TapManagerConfig, shouldUpdateTappers bool) (*K8sTapManager, error) {
	manager := &K8sTapManager{
		context:             ctx,
		CurrentlyTappedPods: make([]core.Pod, 0),
		config:              config,
		kubernetesProvider:  kubernetesProvider,
		TapPodChangesOut:    make(chan TappedPodChangeEvent, 100),
		ErrorOut:            make(chan K8sTapManagerError, 100),
		shouldUpdateTappers: shouldUpdateTappers,
	}

	if err, _ := manager.updateCurrentlyTappedPods(); err != nil {
		return nil, err
	}

	if shouldUpdateTappers {
		if err := manager.updateMizuTappers(); err != nil {
			return nil, err
		}
	}

	go goUtils.HandleExcWrapper(manager.watchPodsForTapping)
	return manager, nil
}

// BeginUpdatingTappers should only be called after mizu api server is available
func (tapManager *K8sTapManager) BeginUpdatingTappers() error {
	tapManager.shouldUpdateTappers = true
	if err := tapManager.updateMizuTappers(); err != nil {
		return err
	}
	return nil
}

func (tapManager *K8sTapManager) watchPodsForTapping() {
	added, modified, removed, errorChan := FilteredWatch(tapManager.context, tapManager.kubernetesProvider, tapManager.config.TargetNamespaces, &tapManager.config.PodFilterRegex)

	restartTappers := func() {
		err, changeFound := tapManager.updateCurrentlyTappedPods()
		if err != nil {
			tapManager.ErrorOut <- K8sTapManagerError{
				OriginalError:    err,
				TapManagerReason: TapManagerPodListError,
			}
		}

		if !changeFound {
			logger.Log.Debugf("Nothing changed update tappers not needed")
			return
		}
		if tapManager.shouldUpdateTappers {
			if err := tapManager.updateMizuTappers(); err != nil {
				tapManager.ErrorOut <- K8sTapManagerError{
					OriginalError:    err,
					TapManagerReason: TapManagerTapperUpdateError,
				}
			}
		}
	}
	restartTappersDebouncer := debounce.NewDebouncer(updateTappersDelay, restartTappers)

	for {
		select {
		case pod, ok := <-added:
			if !ok {
				added = nil
				continue
			}

			logger.Log.Debugf("Added matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case pod, ok := <-removed:
			if !ok {
				removed = nil
				continue
			}

			logger.Log.Debugf("Removed matching pod %s, ns: %s", pod.Name, pod.Namespace)
			restartTappersDebouncer.SetOn()
		case pod, ok := <-modified:
			if !ok {
				modified = nil
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

			logger.Log.Debugf("Watching pods loop, got error %v, stopping `restart tappers debouncer`", err)
			restartTappersDebouncer.Cancel()
			tapManager.ErrorOut <- K8sTapManagerError{
				OriginalError:    err,
				TapManagerReason: TapManagerPodWatchError,
			}
			// TODO: Does this also perform cleanup?

		case <-tapManager.context.Done():
			logger.Log.Debugf("Watching pods loop, context done, stopping `restart tappers debouncer`")
			restartTappersDebouncer.Cancel()
			return
		}
	}
}

func (tapManager *K8sTapManager) updateCurrentlyTappedPods() (err error, changesFound bool) {
	if matchingPods, err := tapManager.kubernetesProvider.ListAllRunningPodsMatchingRegex(tapManager.context, &tapManager.config.PodFilterRegex, tapManager.config.TargetNamespaces); err != nil {
		return err, false
	} else {
		podsToTap := excludeMizuPods(matchingPods)
		addedPods, removedPods := getPodArrayDiff(tapManager.CurrentlyTappedPods, podsToTap)
		if len(addedPods) > 0 || len(removedPods) > 0 {
			tapManager.CurrentlyTappedPods = podsToTap
			tapManager.TapPodChangesOut <- TappedPodChangeEvent{
				Added:   addedPods,
				Removed: removedPods,
			}
			return nil, true
		}
		return nil, false
	}
}

func (tapManager *K8sTapManager) updateMizuTappers() error {
	nodeToTappedPodIPMap := GetNodeHostToTappedPodIpsMap(tapManager.CurrentlyTappedPods)

	if len(nodeToTappedPodIPMap) > 0 {
		var serviceAccountName string
		if tapManager.config.MizuServiceAccountExists {
			serviceAccountName = ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		if err := tapManager.kubernetesProvider.ApplyMizuTapperDaemonSet(
			tapManager.context,
			tapManager.config.MizuResourcesNamespace,
			TapperDaemonSetName,
			tapManager.config.AgentImage,
			TapperPodName,
			fmt.Sprintf("%s.%s.svc.cluster.local", ApiServerPodName, tapManager.config.MizuResourcesNamespace),
			nodeToTappedPodIPMap,
			serviceAccountName,
			tapManager.config.TapperResources,
			tapManager.config.ImagePullPolicy,
			tapManager.config.MizuApiFilteringOptions,
			tapManager.config.DumpLogs,
		); err != nil {
			return err
		}
		logger.Log.Debugf("Successfully created %v tappers", len(nodeToTappedPodIPMap))
	} else {
		if err := tapManager.kubernetesProvider.RemoveDaemonSet(tapManager.context, tapManager.config.MizuResourcesNamespace, TapperDaemonSetName); err != nil {
			return err
		}
	}

	return nil
}
