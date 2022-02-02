package tapConfig

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/config"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/agent/pkg/providers/database"
	"github.com/up9inc/mizu/agent/pkg/providers/tappedPods"
	"github.com/up9inc/mizu/agent/pkg/providers/tappers"
	"github.com/up9inc/mizu/agent/pkg/utils"

	kubernetesProvider "github.com/up9inc/mizu/agent/pkg/providers/kubernetes"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
	v1 "k8s.io/api/core/v1"
)

const FilePath = shared.DataDirPath + "tap-config.json"

var (
	lock               = &sync.Mutex{}
	syncOnce           sync.Once
	tapConfig          *models.TapConfig
	cancelTapperSyncer context.CancelFunc
)

func Get() *models.TapConfig {
	syncOnce.Do(func() {
		if err := utils.ReadJsonFile(FilePath, &tapConfig); err != nil {
			tapConfig = &models.TapConfig{TappedNamespaces: make(map[string]bool)}

			if !os.IsNotExist(err) {
				logger.Log.Errorf("Error reading tap config from file, err: %v", err)
			}
		}
	})

	return tapConfig
}

func Save(tapConfigToSave *models.TapConfig) {
	lock.Lock()
	defer lock.Unlock()

	tapConfig = tapConfigToSave
	if err := utils.SaveJsonFile(FilePath, tapConfig); err != nil {
		logger.Log.Errorf("Error saving tap config, err: %v", err)
	}
}

// this function fetches the union of all configured namespaces in all workspaces and configures tapperSyncer to tap those namespaces
func SyncTappingConfigWithWorkspaceNamespaces() error {
	namespacesToTap, err := database.GetAllUniqueNamespaces()

	if err != nil {
		return err
	}

	if cancelTapperSyncer != nil {
		cancelTapperSyncer()

		tappedPods.Set([]*shared.PodInfo{})
		tappers.ResetStatus()

		broadcastTappedPodsStatusPerWorkspace()
	}

	podRegex, _ := regexp.Compile(".*")

	kubernetesProvider, err := kubernetesProvider.GetKubernetesProvider()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	if _, err := startMizuTapperSyncer(ctx, kubernetesProvider, namespacesToTap, *podRegex, []string{}, tapApi.TrafficFilteringOptions{}, false); err != nil {
		cancel()
		return err
	} else {
		cancelTapperSyncer = cancel
		return nil
	}
}

func startMizuTapperSyncer(ctx context.Context, provider *kubernetes.Provider, targetNamespaces []string, podFilterRegex regexp.Regexp, ignoredUserAgents []string, mizuApiFilteringOptions tapApi.TrafficFilteringOptions, serviceMesh bool) (*kubernetes.MizuTapperSyncer, error) {
	tapperSyncer, err := kubernetes.CreateAndStartMizuTapperSyncer(ctx, provider, kubernetes.TapperSyncerConfig{
		TargetNamespaces:         targetNamespaces,
		PodFilterRegex:           podFilterRegex,
		MizuResourcesNamespace:   config.Config.MizuResourcesNamespace,
		AgentImage:               config.Config.AgentImage,
		TapperResources:          config.Config.TapperResources,
		ImagePullPolicy:          v1.PullPolicy(config.Config.PullPolicy),
		LogLevel:                 config.Config.LogLevel,
		IgnoredUserAgents:        ignoredUserAgents,
		MizuApiFilteringOptions:  mizuApiFilteringOptions,
		MizuServiceAccountExists: true, //assume service account exists since install mode will not function without it anyway
		ServiceMesh:              serviceMesh,
	}, time.Now())

	if err != nil {
		return nil, err
	}

	// handle tapperSyncer events (pod changes and errors)
	go func() {
		for {
			select {
			case syncerErr, ok := <-tapperSyncer.ErrorOut:
				if !ok {
					logger.Log.Debug("mizuTapperSyncer err channel closed, ending listener loop")
					return
				}
				logger.Log.Fatalf("fatal tap syncer error: %v", syncerErr)
			case _, ok := <-tapperSyncer.TapPodChangesOut:
				if !ok {
					logger.Log.Debug("mizuTapperSyncer pod changes channel closed, ending listener loop")
					return
				}

				tappedPods.Set(kubernetes.GetPodInfosForPods(tapperSyncer.CurrentlyTappedPods))
				broadcastTappedPodsStatusPerWorkspace()
			case tapperStatus, ok := <-tapperSyncer.TapperStatusChangedOut:
				if !ok {
					logger.Log.Debug("mizuTapperSyncer tapper status changed channel closed, ending listener loop")
					return
				}

				tappers.SetStatus(&tapperStatus)
				broadcastTappedPodsStatusPerWorkspace()
			case <-ctx.Done():
				logger.Log.Debug("mizuTapperSyncer event listener loop exiting due to context done")
				return
			}
		}
	}()

	return tapperSyncer, nil
}

func broadcastTappedPodsStatusPerWorkspace() {
	tappedPodsStatus := tappedPods.GetTappedPodsStatus()
	workspaces, err := database.ListWorkspacesWithRelations()
	if err != nil {
		logger.Log.Errorf("Error listing workspaces for tap pod status update %v", err)
	}

	for _, workspace := range workspaces {
		workspaceTapStatus := filterTapStatusByNamespaceObjects(tappedPodsStatus, workspace.Namespaces)
		message := shared.CreateWebSocketStatusMessage(workspaceTapStatus)
		if jsonBytes, err := json.Marshal(message); err != nil {
			logger.Log.Errorf("Could not Marshal message %v", err)
		} else {
			logger.Log.Infof("broadcasting for workspace")
			api.BroadcastToBrowserClientsByTopic(jsonBytes, &workspace.Id)
		}
	}
}

func filterTapStatusByNamespaceObjects(tappedPodStatus []shared.TappedPodStatus, namespaceObjects []database.Namespace) []shared.TappedPodStatus {
	namespaces := make([]string, len(namespaceObjects))
	for i, namespaceObject := range namespaceObjects {
		namespaces[i] = namespaceObject.Name
	}

	return utils.FilterTapStatusByNamespaces(tappedPodStatus, namespaces)
}
