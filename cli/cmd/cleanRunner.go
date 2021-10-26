package cmd

import (
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
)

func performCleanCommand() {
	kubernetesProvider, err := kubernetes.NewProvider(config.Config.KubeConfigPath())
	if err != nil {
		logger.Log.Error(err)
		return
	}

	finishMizuExecution(kubernetesProvider)
}