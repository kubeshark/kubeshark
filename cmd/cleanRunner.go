package cmd

import (
	"github.com/kubeshark/kubeshark/config"
)

func performCleanCommand() {
	kubernetesProvider, err := getKubernetesProviderForCli(false, false)
	if err != nil {
		return
	}

	finishSelfExecution(kubernetesProvider, config.Config.IsNsRestrictedMode(), config.Config.Tap.SelfNamespace, false)
}
