package check

import (
	"context"
	"fmt"
	"regexp"

	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared/kubernetes"
)

func ServerConnection(kubernetesProvider *kubernetes.Provider) bool {
	logger.Log.Infof("\nAPI-server-connectivity\n--------------------")

	serverUrl := fmt.Sprintf("http://%s", kubernetes.GetMizuApiServerProxiedHostAndPath(config.Config.Tap.GuiPort))

	apiServerProvider := apiserver.NewProvider(serverUrl, 1, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(); err == nil {
		logger.Log.Infof("%v found Mizu server tunnel available and connected successfully to API server", fmt.Sprintf(uiUtils.Green, "√"))
		return true
	}

	connectedToApiServer := false

	if err := checkProxy(serverUrl, kubernetesProvider); err != nil {
		logger.Log.Errorf("%v couldn't connect to API server using proxy, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
	} else {
		connectedToApiServer = true
		logger.Log.Infof("%v connected successfully to API server using proxy", fmt.Sprintf(uiUtils.Green, "√"))
	}

	if err := checkPortForward(serverUrl, kubernetesProvider); err != nil {
		logger.Log.Errorf("%v couldn't connect to API server using port-forward, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
	} else {
		connectedToApiServer = true
		logger.Log.Infof("%v connected successfully to API server using port-forward", fmt.Sprintf(uiUtils.Green, "√"))
	}

	return connectedToApiServer
}

func checkProxy(serverUrl string, kubernetesProvider *kubernetes.Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, err := kubernetes.StartProxy(kubernetesProvider, config.Config.Tap.ProxyHost, config.Config.Tap.GuiPort, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName, cancel)
	if err != nil {
		return err
	}

	apiServerProvider := apiserver.NewProvider(serverUrl, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(); err != nil {
		return err
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Log.Debugf("Error occurred while stopping proxy, err: %v", err)
	}

	return nil
}

func checkPortForward(serverUrl string, kubernetesProvider *kubernetes.Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podRegex, _ := regexp.Compile(kubernetes.ApiServerPodName)
	forwarder, err := kubernetes.NewPortForward(kubernetesProvider, config.Config.MizuResourcesNamespace, podRegex, config.Config.Tap.GuiPort, ctx, cancel)
	if err != nil {
		return err
	}

	apiServerProvider := apiserver.NewProvider(serverUrl, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	if err := apiServerProvider.TestConnection(); err != nil {
		return err
	}

	forwarder.Close()

	return nil
}
