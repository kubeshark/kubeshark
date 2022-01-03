package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
	v1 "k8s.io/api/core/v1"
	"mizuserver/pkg/config"
	"mizuserver/pkg/models"
	"mizuserver/pkg/providers"
	"net/http"
	"regexp"
	"time"
)

var globalStandaloneConfig *models.StandaloneConfig
var cancelTapperSyncer context.CancelFunc

func UpdateConfig(c *gin.Context) {
	standaloneConfig := &models.StandaloneConfig{}

	if err := c.Bind(standaloneConfig); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	podRegex, err := regexp.Compile(standaloneConfig.PodRegex)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if cancelTapperSyncer != nil {
		cancelTapperSyncer()

		providers.TapStatus = shared.TapStatus{}
		providers.TappersStatus = make(map[string]shared.TapperStatus)

		broadcastTappedPodsStatus()
	}

	ctx, cancel := context.WithCancel(context.Background())

	kubernetesProvider, err := kubernetes.NewProviderInCluster()
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		cancel()
		return
	}

	if _, err = startMizuTapperSyncer(ctx, kubernetesProvider, standaloneConfig.TargetNamespaces, *podRegex, []string{} , tapApi.TrafficFilteringOptions{}, false); err != nil {
		c.JSON(http.StatusBadRequest, err)
		cancel()
		return
	}

	cancelTapperSyncer = cancel
	globalStandaloneConfig = standaloneConfig

	c.JSON(http.StatusOK, "OK")
}

func GetConfig(c *gin.Context) {
	if globalStandaloneConfig != nil {
		c.JSON(http.StatusOK, globalStandaloneConfig)
	}
}

func startMizuTapperSyncer(ctx context.Context, provider *kubernetes.Provider, targetNamespaces []string, podFilterRegex regexp.Regexp, ignoredUserAgents []string, mizuApiFilteringOptions tapApi.TrafficFilteringOptions, istio bool) (*kubernetes.MizuTapperSyncer, error) {
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
		Istio:                    istio,
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

				providers.TapStatus = shared.TapStatus{Pods: kubernetes.GetPodInfosForPods(tapperSyncer.CurrentlyTappedPods)}
				broadcastTappedPodsStatus()
			case tapperStatus, ok := <-tapperSyncer.TapperStatusChangedOut:
				if !ok {
					logger.Log.Debug("mizuTapperSyncer tapper status changed channel closed, ending listener loop")
					return
				}

				addTapperStatus(tapperStatus)
				broadcastTappedPodsStatus()
			case <-ctx.Done():
				logger.Log.Debug("mizuTapperSyncer event listener loop exiting due to context done")
				return
			}
		}
	}()

	return tapperSyncer, nil
}
