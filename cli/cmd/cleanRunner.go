package cmd

import (
	"github.com/kubeshark/kubeshark/cli/config"
)

func performCleanCommand() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	finishKubesharkExecution(kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.KubesharkResourcesNamespace)
}
