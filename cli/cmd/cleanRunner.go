package cmd

import "github.com/up9inc/mizu/cli/apiserver"

func performCleanCommand() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	finishMizuExecution(kubernetesProvider, apiserver.NewProvider(GetApiServerUrl(), apiserver.DefaultRetries, apiserver.DefaultTimeout))
}
