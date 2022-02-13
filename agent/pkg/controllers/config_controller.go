package controllers

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/config"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/agent/pkg/providers"
	"github.com/up9inc/mizu/agent/pkg/providers/tapConfig"
	"github.com/up9inc/mizu/agent/pkg/providers/tappedPods"
	"github.com/up9inc/mizu/agent/pkg/providers/tappers"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
	v1 "k8s.io/api/core/v1"
)

var cancelTapperSyncer context.CancelFunc

func PostTapConfig(c *gin.Context) {
	requestTapConfig := &models.TapConfig{}

	if err := c.Bind(requestTapConfig); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if cancelTapperSyncer != nil {
		cancelTapperSyncer()

		tappedPods.Set([]*shared.PodInfo{})
		tappers.ResetStatus()

		api.BroadcastTappedPodsStatus()
	}

	var tappedNamespaces []string
	for namespace, tapped := range requestTapConfig.TappedNamespaces {
		if tapped {
			tappedNamespaces = append(tappedNamespaces, namespace)
		}
	}

	podRegex, _ := regexp.Compile(".*")

	kubernetesProvider, err := providers.GetKubernetesProvider()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	if _, err := startMizuTapperSyncer(ctx, kubernetesProvider, tappedNamespaces, *podRegex, []string{}, tapApi.TrafficFilteringOptions{}, false, false); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		cancel()
		return
	}

	cancelTapperSyncer = cancel
	tapConfig.Save(requestTapConfig)

	c.JSON(http.StatusOK, "OK")
}

func GetTapConfig(c *gin.Context) {
	kubernetesProvider, err := providers.GetKubernetesProvider()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	namespaces, err := kubernetesProvider.ListAllNamespaces(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	savedTapConfig := tapConfig.Get()

	tappedNamespaces := make(map[string]bool)
	for _, namespace := range namespaces {
		if namespace.Name == config.Config.MizuResourcesNamespace {
			continue
		}

		tappedNamespaces[namespace.Name] = savedTapConfig.TappedNamespaces[namespace.Name]
	}

	tapConfigToReturn := models.TapConfig{TappedNamespaces: tappedNamespaces}
	c.JSON(http.StatusOK, tapConfigToReturn)
}

func startMizuTapperSyncer(ctx context.Context, provider *kubernetes.Provider, targetNamespaces []string, podFilterRegex regexp.Regexp, ignoredUserAgents []string, mizuApiFilteringOptions tapApi.TrafficFilteringOptions, serviceMesh bool, tls bool) (*kubernetes.MizuTapperSyncer, error) {
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
		Tls:                      tls,
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
				api.BroadcastTappedPodsStatus()
			case tapperStatus, ok := <-tapperSyncer.TapperStatusChangedOut:
				if !ok {
					logger.Log.Debug("mizuTapperSyncer tapper status changed channel closed, ending listener loop")
					return
				}

				tappers.SetStatus(&tapperStatus)
				api.BroadcastTappedPodsStatus()
			case <-ctx.Done():
				logger.Log.Debug("mizuTapperSyncer event listener loop exiting due to context done")
				return
			}
		}
	}()

	return tapperSyncer, nil
}
