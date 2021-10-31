package cmd

func performCleanCommand() {
	kubernetesProvider, err := getKubernetesProviderForCli()
	if err != nil {
		return
	}

	finishMizuExecution(kubernetesProvider)
}
