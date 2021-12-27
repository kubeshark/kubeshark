package cmd

import (
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
)

func performCleanCommand() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	finishMizuExecution(kubernetesProvider, apiserver.NewProvider(GetApiServerUrl(), apiserver.DefaultRetries, apiserver.DefaultTimeout), config.Config.IsNsRestrictedMode(), config.Config.MizuResourcesNamespace)
}
