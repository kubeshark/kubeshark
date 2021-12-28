package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/resources"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
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

	logger.Log.Infof("Waiting for Mizu Agent to start...")
	if state.mizuServiceAccountExists, err = resources.CreateMizuResources(ctx, cancel, kubernetesProvider, serializedValidationRules, serializedContract, serializedMizuConfig, config.Config.IsNsRestrictedMode(), config.Config.MizuResourcesNamespace, true, config.Config.AgentImage, nil, defaultMaxEntriesDBSizeBytes, defaultResources, config.Config.ImagePullPolicy(), config.Config.LogLevel(), false); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error creating resources: %v", errormessage.FormatError(err)))

		var statusError *k8serrors.StatusError
		if errors.As(err, &statusError) {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				logger.Log.Info("Mizu is already running in this namespace, change the `mizu-resources-namespace` configuration or run `mizu clean` to remove the currently running Mizu instance")
			}
		}
		return
	}

	if err := handleInstallModePostCreation(cancel, kubernetesProvider); err != nil {
		defer finishMizuExecution(kubernetesProvider, apiProvider, config.Config.IsNsRestrictedMode(), config.Config.MizuResourcesNamespace)
		cancel()
	} else {
		logger.Log.Infof(uiUtils.Magenta, "Mizu is now running in install mode, run `mizu view` to connect to the mizu daemon instance")
	}
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
	}

	return &mizuAgentConfig
}

func handleInstallModePostCreation(cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) error {
	apiProvider := apiserver.NewProvider(GetApiServerUrl(), 90, 1*time.Second)

	if err := waitForInstallModeToBeReady(cancel, kubernetesProvider, apiProvider); err != nil {
		return err
	}

	return nil
}

func waitForInstallModeToBeReady(cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, apiProvider *apiserver.Provider) error {
	go startProxyReportErrorIfAny(kubernetesProvider, cancel)

	if err := apiProvider.TestConnection(); err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Mizu was not ready in time, for more info check logs at %s", fsUtils.GetLogFilePath()))
		return err
	}
	return nil
}
