package cmd

import (
	"github.com/up9inc/mizu/cli/config"
)

func performCleanCommand() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	finishMizuExecution(kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.MizuResourcesNamespace)
}
