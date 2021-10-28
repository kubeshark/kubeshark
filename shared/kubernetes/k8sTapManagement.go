package kubernetes

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	core "k8s.io/api/core/v1"
)

type K8sTapManager struct {
	state TapState
	Config TapManagerConfig
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
			tapManager.Config.MizuResourcesNamespace,
			TapperDaemonSetName,
			tapManager.Config.AgentImage,
			TapperPodName,
			fmt.Sprintf("%s.%s.svc.cluster.local", tapManager.state.apiServerService.Name, tapManager.state.apiServerService.Namespace),
			nodeToTappedPodIPMap,
			serviceAccountName,
			tapManager.Config.TapperResources,
			tapManager.Config.ImagePullPolicy,
			mizuApiFilteringOptions,
			tapManager.Config.DumpLogs,
		); err != nil {
			return err
		}
		logger.Log.Debugf("Successfully created %v tappers", len(nodeToTappedPodIPMap))
	} else {
		if err := kubernetesProvider.RemoveDaemonSet(ctx, tapManager.Config.MizuResourcesNamespace, TapperDaemonSetName); err != nil {
			return err
		}
	}

	return nil
}