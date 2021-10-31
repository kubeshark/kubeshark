package cmd

import (
	"github.com/up9inc/mizu/shared/logger"
)

func performCleanCommand() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		logger.Log.Error(err)
		return
	}

	finishMizuExecution(kubernetesProvider)
}
