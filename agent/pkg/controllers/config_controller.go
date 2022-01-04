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

var globalTapConfig *models.TapConfig
var cancelTapperSyncer context.CancelFunc
var kubernetesProvider *kubernetes.Provider

func PostTapConfig(c *gin.Context) {
	tapConfig := &models.TapConfig{}

	if err := c.Bind(tapConfig); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if cancelTapperSyncer != nil {
		cancelTapperSyncer()

		providers.TapStatus = shared.TapStatus{}
		providers.TappersStatus = make(map[string]shared.TapperStatus)

		broadcastTappedPodsStatus()
	}

	if kubernetesProvider == nil {
		var err error
		kubernetesProvider, err = kubernetes.NewProviderInCluster()
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	var tappedNamespaces []string
	for namespace, tapped := range tapConfig.TappedNamespaces {
		if tapped {
			tappedNamespaces = append(tappedNamespaces, namespace)
		}
	}

	podRegex, _ := regexp.Compile(".*")

	if _, err := startMizuTapperSyncer(ctx, kubernetesProvider, tappedNamespaces, *podRegex, []string{} , tapApi.TrafficFilteringOptions{}, false); err != nil {
		c.JSON(http.StatusBadRequest, err)
		cancel()
		return
	}

	cancelTapperSyncer = cancel
	globalTapConfig = tapConfig

	c.JSON(http.StatusOK, "OK")
}

func GetTapConfig(c *gin.Context) {
	if globalTapConfig != nil {
		c.JSON(http.StatusOK, globalTapConfig)
	}

	c.JSON(http.StatusBadRequest, "Not config found")
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
