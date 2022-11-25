package cmd

import (
	"github.com/kubeshark/kubeshark/config"
)

func performCleanCommand() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	finishKubesharkExecution(kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.KubesharkResourcesNamespace)
}
