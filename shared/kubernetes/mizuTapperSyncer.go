package kubernetes

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/op/go-logging"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	"github.com/up9inc/mizu/tap/api"
	core "k8s.io/api/core/v1"
)

const updateTappersDelay = 5 * time.Second

type TappedPodChangeEvent struct {
	Added   []core.Pod
	Removed []core.Pod
}

// MizuTapperSyncer uses a k8s pod watch to update tapper daemonsets when targeted pods are removed or created
type MizuTapperSyncer struct {
	startTime              time.Time
	context                context.Context
	CurrentlyTappedPods    []core.Pod
	config                 TapperSyncerConfig
	kubernetesProvider     *Provider
	TapPodChangesOut       chan TappedPodChangeEvent
	TapperStatusChangedOut chan shared.TapperStatus
	ErrorOut               chan K8sTapManagerError
	nodeToTappedPodMap     shared.NodeToPodsMap
	tappedNodes            []string
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
	ServiceMesh              bool
	Tls                      bool
}

func CreateAndStartMizuTapperSyncer(ctx context.Context, kubernetesProvider *Provider, config TapperSyncerConfig, startTime time.Time) (*MizuTapperSyncer, error) {
	syncer := &MizuTapperSyncer{
		startTime:              startTime.Truncate(time.Second), // Round down because k8s CreationTimestamp is given in 1 sec resolution.
		context:                ctx,
		CurrentlyTappedPods:    make([]core.Pod, 0),
		config:                 config,
		kubernetesProvider:     kubernetesProvider,
		TapPodChangesOut:       make(chan TappedPodChangeEvent, 100),
		TapperStatusChangedOut: make(chan shared.TapperStatus, 100),
		ErrorOut:               make(chan K8sTapManagerError, 100),
	}

	if err, _ := syncer.updateCurrentlyTappedPods(); err != nil {
		return nil, err
	}

	if err := syncer.updateMizuTappers(); err != nil {
		return nil, err
	}

	go syncer.watchPodsForTapping()
	go syncer.watchTapperEvents()
	go syncer.watchTapperPods()
	return syncer, nil
}

func (tapperSyncer *MizuTapperSyncer) watchTapperPods() {
	mizuResourceRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", TapperPodName))
	podWatchHelper := NewPodWatchHelper(tapperSyncer.kubernetesProvider, mizuResourceRegex)
	eventChan, errorChan := FilteredWatch(tapperSyncer.context, podWatchHelper, []string{tapperSyncer.config.MizuResourcesNamespace}, podWatchHelper)

	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			pod, err := wEvent.ToPod()
			if err != nil {
				logger.Log.Debugf("[ERROR] parsing Mizu resource pod: %+v", err)
				continue
			}

			logger.Log.Debugf("Watching tapper pods loop, tapper: %v, node: %v, status: %v", pod.Name, pod.Spec.NodeName, pod.Status.Phase)
			if pod.Spec.NodeName != "" {
				tapperStatus := shared.TapperStatus{TapperName: pod.Name, NodeName: pod.Spec.NodeName, Status: string(pod.Status.Phase)}
				tapperSyncer.TapperStatusChangedOut <- tapperStatus
			}

		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}
			logger.Log.Debugf("[ERROR] Watching tapper pods loop, error: %+v", err)

		case <-tapperSyncer.context.Done():
			logger.Log.Debugf("Watching tapper pods loop, ctx done")
			return
		}
	}
}

func (tapperSyncer *MizuTapperSyncer) watchTapperEvents() {
	mizuResourceRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", TapperPodName))
	eventWatchHelper := NewEventWatchHelper(tapperSyncer.kubernetesProvider, mizuResourceRegex, "pod")
	eventChan, errorChan := FilteredWatch(tapperSyncer.context, eventWatchHelper, []string{tapperSyncer.config.MizuResourcesNamespace}, eventWatchHelper)

	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			event, err := wEvent.ToEvent()
			if err != nil {
				logger.Log.Debugf("[ERROR] parsing Mizu resource event: %+v", err)
				continue
			}

			logger.Log.Debugf(
				fmt.Sprintf("Watching tapper events loop, event %s, time: %v, resource: %s (%s), reason: %s, note: %s",
					event.Name,
					event.CreationTimestamp.Time,
					event.Regarding.Name,
					event.Regarding.Kind,
					event.Reason,
					event.Note))

			pod, err1 := tapperSyncer.kubernetesProvider.GetPod(tapperSyncer.context, tapperSyncer.config.MizuResourcesNamespace, event.Regarding.Name)
			if err1 != nil {
				logger.Log.Debugf(fmt.Sprintf("Couldn't get tapper pod %s", event.Regarding.Name))
				continue
			}

			nodeName := ""
			if event.Reason != "FailedScheduling" {
				nodeName = pod.Spec.NodeName
			} else {
				nodeName = pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchFields[0].Values[0]
			}

			tapperStatus := shared.TapperStatus{TapperName: pod.Name, NodeName: nodeName, Status: string(pod.Status.Phase)}
			tapperSyncer.TapperStatusChangedOut <- tapperStatus

		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			logger.Log.Debugf("[ERROR] Watching tapper events loop, error: %+v", err)

		case <-tapperSyncer.context.Done():
			logger.Log.Debugf("Watching tapper events loop, ctx done")
			return
		}
	}
}

func (tapperSyncer *MizuTapperSyncer) watchPodsForTapping() {
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
				logger.Log.Debugf("Added matching pod %s, ns: %s", pod.Name, pod.Namespace)
				if err := restartTappersDebouncer.SetOn(); err != nil {
					logger.Log.Error(err)
				}
			case EventDeleted:
				logger.Log.Debugf("Removed matching pod %s, ns: %s", pod.Name, pod.Namespace)
				if err := restartTappersDebouncer.SetOn(); err != nil {
					logger.Log.Error(err)
				}
			case EventModified:
				logger.Log.Debugf("Modified matching pod %s, ns: %s, phase: %s, ip: %s", pod.Name, pod.Namespace, pod.Status.Phase, pod.Status.PodIP)
				// Act only if the modified pod has already obtained an IP address.
				// After filtering for IPs, on a normal pod restart this includes the following events:
				// - Pod deletion
				// - Pod reaches start state
				// - Pod reaches ready state
				// Ready/unready transitions might also trigger this event.
				if pod.Status.PodIP != "" {
					if err := restartTappersDebouncer.SetOn(); err != nil {
						logger.Log.Error(err)
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

func (tapperSyncer *MizuTapperSyncer) updateMizuTappers() error {
	nodesToTap := make([]string, len(tapperSyncer.nodeToTappedPodMap))
	i := 0
	for node := range tapperSyncer.nodeToTappedPodMap {
		nodesToTap[i] = node
		i++
	}

	if shared.EqualStringSlices(nodesToTap, tapperSyncer.tappedNodes) {
		logger.Log.Debug("Skipping apply, DaemonSet is up to date")
		return nil
	}

	logger.Log.Debugf("Updating DaemonSet to run on nodes: %v", nodesToTap)

	if len(tapperSyncer.nodeToTappedPodMap) > 0 {
		var serviceAccountName string
		if tapperSyncer.config.MizuServiceAccountExists {
			serviceAccountName = ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		nodeNames := make([]string, 0, len(tapperSyncer.nodeToTappedPodMap))
		for nodeName := range tapperSyncer.nodeToTappedPodMap {
			nodeNames = append(nodeNames, nodeName)
		}

		if err := tapperSyncer.kubernetesProvider.ApplyMizuTapperDaemonSet(
			tapperSyncer.context,
			tapperSyncer.config.MizuResourcesNamespace,
			TapperDaemonSetName,
			tapperSyncer.config.AgentImage,
			TapperPodName,
			fmt.Sprintf("%s.%s.svc", ApiServerPodName, tapperSyncer.config.MizuResourcesNamespace),
			nodeNames,
			serviceAccountName,
			tapperSyncer.config.TapperResources,
			tapperSyncer.config.ImagePullPolicy,
			tapperSyncer.config.MizuApiFilteringOptions,
			tapperSyncer.config.LogLevel,
			tapperSyncer.config.ServiceMesh,
			tapperSyncer.config.Tls); err != nil {
			return err
		}

		logger.Log.Debugf("Successfully created %v tappers", len(tapperSyncer.nodeToTappedPodMap))
	} else {
		if err := tapperSyncer.kubernetesProvider.ResetMizuTapperDaemonSet(
			tapperSyncer.context,
			tapperSyncer.config.MizuResourcesNamespace,
			TapperDaemonSetName,
			tapperSyncer.config.AgentImage,
			TapperPodName); err != nil {
			return err
		}

		logger.Log.Debugf("Successfully reset tapper daemon set")
	}

	tapperSyncer.tappedNodes = nodesToTap

	return nil
}
