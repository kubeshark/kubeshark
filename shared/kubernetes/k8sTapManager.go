package kubernetes

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	core "k8s.io/api/core/v1"
	"regexp"
)

type K8sTapManager struct {
	state TapState
	config TapManagerConfig
	podFilterRegex regexp.Regexp
}

type TapState struct {
	apiServerService         *core.Service
	currentlyTappedPods      []core.Pod
	mizuServiceAccountExists bool
}

type TapManagerConfig struct {
	MizuResourcesNamespace string
	AgentImage            string
	TapperResources       shared.Resources
	ImagePullPolicy       core.PullPolicy
	DumpLogs              bool
	IgnoredUserAgents     []string
}

func CreateK8sTapManager(config TapManagerConfig, podFilterRegex regexp.Regexp) *K8sTapManager {
	return &K8sTapManager{
		state:          TapState{},
		config:         config,
		podFilterRegex: podFilterRegex,
	}
}

func updateCurrentlyTappedPods(kubernetesProvider *Provider, ctx context.Context, targetNamespaces []string) (error, bool) {
	changeFound := false
	if matchingPods, err := kubernetesProvider.ListAllRunningPodsMatchingRegex(ctx, config.Config.Tap.PodRegex(), targetNamespaces); err != nil {
		return err, false
	} else {
		podsToTap := excludeMizuPods(matchingPods)
		addedPods, removedPods := getPodArrayDiff(state.currentlyTappedPods, podsToTap)
		for _, addedPod := range addedPods {
			changeFound = true
			logger.Log.Infof(uiUtils.Green, fmt.Sprintf("+%s", addedPod.Name))
		}
		for _, removedPod := range removedPods {
			changeFound = true
			logger.Log.Infof(uiUtils.Red, fmt.Sprintf("-%s", removedPod.Name))
		}
		state.currentlyTappedPods = podsToTap
	}

	return nil, changeFound
}

func (tapManager *K8sTapManager) updateMizuTappers(ctx context.Context, kubernetesProvider *Provider, mizuApiFilteringOptions *interface{}) error {
	nodeToTappedPodIPMap := GetNodeHostToTappedPodIpsMap(tapManager.state.currentlyTappedPods)

	if len(nodeToTappedPodIPMap) > 0 {
		var serviceAccountName string
		if tapManager.state.mizuServiceAccountExists {
			serviceAccountName = ServiceAccountName
		} else {
			serviceAccountName = ""
		}

		if err := kubernetesProvider.ApplyMizuTapperDaemonSet(
			ctx,
			tapManager.config.MizuResourcesNamespace,
			TapperDaemonSetName,
			tapManager.config.AgentImage,
			TapperPodName,
			fmt.Sprintf("%s.%s.svc.cluster.local", tapManager.state.apiServerService.Name, tapManager.state.apiServerService.Namespace),
			nodeToTappedPodIPMap,
			serviceAccountName,
			tapManager.config.TapperResources,
			tapManager.config.ImagePullPolicy,
			mizuApiFilteringOptions,
			tapManager.config.DumpLogs,
		); err != nil {
			return err
		}
		logger.Log.Debugf("Successfully created %v tappers", len(nodeToTappedPodIPMap))
	} else {
		if err := kubernetesProvider.RemoveDaemonSet(ctx, tapManager.config.MizuResourcesNamespace, TapperDaemonSetName); err != nil {
			return err
		}
	}

	return nil
}