package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/telemetry"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
)

func GetApiServerUrl() string {
	return fmt.Sprintf("http://%s", kubernetes.GetMizuApiServerProxiedHostAndPath(config.Config.Tap.GuiPort))
}

func startProxyReportErrorIfAny(kubernetesProvider *kubernetes.Provider, cancel context.CancelFunc) {
	err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.ProxyHost, config.Config.Tap.GuiPort, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName)
	if err != nil {
		logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error occured while running k8s proxy %v\n"+
			"Try setting different port by using --%s", errormessage.FormatError(err), configStructs.GuiPortTapName))
		cancel()
	}

	logger.Log.Debugf("proxy ended")
}

func waitForFinish(ctx context.Context, cancel context.CancelFunc) {
	logger.Log.Debugf("waiting for finish...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// block until ctx cancel is called or termination signal is received
	select {
	case <-ctx.Done():
		logger.Log.Debugf("ctx done")
		break
	case <-sigChan:
		logger.Log.Debugf("Got termination signal, canceling execution...")
		cancel()
	}
}

func getKubernetesProviderForCli() (*kubernetes.Provider, error) {
	kubernetesProvider, err := kubernetes.NewProvider(config.Config.KubeConfigPath())
	if err != nil {
		handleKubernetesProviderError(err)
		return nil, err
	}
	return kubernetesProvider, nil
}

func handleKubernetesProviderError(err error) {
	var clusterBehindProxyErr *kubernetes.ClusterBehindProxyError
	if ok := errors.As(err, &clusterBehindProxyErr); ok {
		logger.Log.Errorf("cannot establish http-proxy connection to the Kubernetes cluster. If you’re using Lens or similar tool, please run mizu with regular kubectl config using --%v %v=$HOME/.kube/config flag", config.SetCommandName, config.KubeConfigPathConfigName)
	} else {
		logger.Log.Error(err)
	}
}

func finishMizuExecution(kubernetesProvider *kubernetes.Provider, apiProvider *apiserver.Provider) {
	telemetry.ReportAPICalls(apiProvider)
	removalCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()
	dumpLogsIfNeeded(removalCtx, kubernetesProvider)
	cleanUpMizuResources(removalCtx, cancel, kubernetesProvider)
}

func dumpLogsIfNeeded(ctx context.Context, kubernetesProvider *kubernetes.Provider) {
	if !config.Config.DumpLogs {
		return
	}
	mizuDir := mizu.GetMizuFolderPath()
	filePath := path.Join(mizuDir, fmt.Sprintf("mizu_logs_%s.zip", time.Now().Format("2006_01_02__15_04_05")))
	if err := fsUtils.DumpLogs(ctx, kubernetesProvider, filePath); err != nil {
		logger.Log.Errorf("Failed dump logs %v", err)
	}
}

func cleanUpMizuResources(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) {
	logger.Log.Infof("\nRemoving mizu resources")

	var leftoverResources []string

	if config.Config.IsNsRestrictedMode() {
		leftoverResources = cleanUpRestrictedMode(ctx, kubernetesProvider)
	} else {
		leftoverResources = cleanUpNonRestrictedMode(ctx, cancel, kubernetesProvider)
	}

	if len(leftoverResources) > 0 {
		errMsg := fmt.Sprintf("Failed to remove the following resources, for more info check logs at %s:", fsUtils.GetLogFilePath())
		for _, resource := range leftoverResources {
			errMsg += "\n- " + resource
		}
		logger.Log.Errorf(uiUtils.Error, errMsg)
	}
}
