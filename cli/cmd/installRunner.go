package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/up9inc/mizu/shared/kubernetes"
	core "k8s.io/api/core/v1"
	"regexp"
	"time"

	"github.com/creasty/defaults"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/resources"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func runMizuInstall() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	var serializedValidationRules string
	var serializedContract string

	var defaultMaxEntriesDBSizeBytes int64 = 200 * 1000 * 1000

	defaultResources := shared.Resources{}
	defaults.Set(&defaultResources)

	mizuAgentConfig := getInstallMizuAgentConfig(defaultMaxEntriesDBSizeBytes, defaultResources)
	serializedMizuConfig, err := getSerializedMizuAgentConfig(mizuAgentConfig)
	if err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error serializing mizu config: %v", errormessage.FormatError(err)))
		return
	}

	if err = resources.CreateInstallMizuResources(ctx, kubernetesProvider, serializedValidationRules,
		serializedContract, serializedMizuConfig, config.Config.IsNsRestrictedMode(),
		config.Config.MizuResourcesNamespace, config.Config.AgentImage,
		nil, defaultMaxEntriesDBSizeBytes, defaultResources, config.Config.ImagePullPolicy(),
		config.Config.LogLevel(), false); err != nil {
		var statusError *k8serrors.StatusError
		if errors.As(err, &statusError) {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				logger.Log.Info("Mizu is already running in this namespace, run `mizu clean` to remove the currently running Mizu instance")
			}
		} else {
			defer resources.CleanUpMizuResources(ctx, cancel, kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.MizuResourcesNamespace)
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error creating resources: %v", errormessage.FormatError(err)))
		}

		return
	}

	logger.Log.Infof("Waiting for Mizu server to start...")
	readyChan := make(chan string)
	readyErrorChan := make(chan error)
	go watchApiServerPodReady(ctx, kubernetesProvider, readyChan, readyErrorChan)

	select {
	case readyMessage := <-readyChan:
		logger.Log.Infof(readyMessage)
	case err := <-readyErrorChan:
		defer resources.CleanUpMizuResources(ctx, cancel, kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.MizuResourcesNamespace)
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("%v", errormessage.FormatError(err)))
		return
	}

	logger.Log.Infof(uiUtils.Magenta, "Installation completed, run `mizu view` to connect to the mizu daemon instance")
}

func getInstallMizuAgentConfig(maxDBSizeBytes int64, tapperResources shared.Resources) *shared.MizuAgentConfig {
	mizuAgentConfig := shared.MizuAgentConfig{
		MaxDBSizeBytes:         maxDBSizeBytes,
		AgentImage:             config.Config.AgentImage,
		PullPolicy:             config.Config.ImagePullPolicyStr,
		LogLevel:               config.Config.LogLevel(),
		TapperResources:        tapperResources,
		MizuResourcesNamespace: config.Config.MizuResourcesNamespace,
		AgentDatabasePath:      shared.DataDirPath,
		StandaloneMode:         true,
	}

	return &mizuAgentConfig
}

func watchApiServerPodReady(ctx context.Context, kubernetesProvider *kubernetes.Provider, readyChan chan string, readyErrorChan chan error) {
	podExactRegex := regexp.MustCompile(fmt.Sprintf("^%s.*", kubernetes.ApiServerPodName))
	podWatchHelper := kubernetes.NewPodWatchHelper(kubernetesProvider, podExactRegex)
	eventChan, errorChan := kubernetes.FilteredWatch(ctx, podWatchHelper, []string{config.Config.MizuResourcesNamespace}, podWatchHelper)

	timeAfter := time.After(1 * time.Minute)
	for {
		select {
		case wEvent, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			switch wEvent.Type {
			case kubernetes.EventAdded:
				logger.Log.Debugf("Watching API Server pod ready loop, added")
			case kubernetes.EventDeleted:
				logger.Log.Debugf("Watching API Server pod ready loop, %s removed", kubernetes.ApiServerPodName)
			case kubernetes.EventModified:
				modifiedPod, err := wEvent.ToPod()
				if err != nil {
					readyErrorChan <- err
					return
				}

				logger.Log.Debugf("Watching API Server pod ready loop, modified: %v", modifiedPod.Status.Phase)

				if modifiedPod.Status.Phase == core.PodRunning {
					readyChan <- fmt.Sprintf("%v pod is running", modifiedPod.Name)
					return
				}
			case kubernetes.EventBookmark:
				break
			case kubernetes.EventError:
				break
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}

			readyErrorChan <- fmt.Errorf("[ERROR] Agent creation, watching %v namespace, error: %v", config.Config.MizuResourcesNamespace, err)
			return
		case <-timeAfter:
			readyErrorChan <- fmt.Errorf("mizu API server was not ready in time")
			return
		case <-ctx.Done():
			logger.Log.Debugf("Watching API Server pod ready loop, ctx done")
			return
		}
	}
}
