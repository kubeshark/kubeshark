package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/kubeshark/kubeshark/cmd/check"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
)

var (
	//go:embed permissionFiles
	embedFS embed.FS
)

func runKubesharkCheck() {
	log.Info().Msg("Checking the deployment...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	kubernetesProvider, kubernetesVersion, checkPassed := check.KubernetesApi()

	if checkPassed {
		checkPassed = check.KubernetesVersion(kubernetesVersion)
	}

	if checkPassed {
		checkPassed = check.KubernetesPermissions(ctx, embedFS, kubernetesProvider)
	}

	if checkPassed {
		checkPassed = check.ImagePullInCluster(ctx, kubernetesProvider)
	}
	if checkPassed {
		checkPassed = check.KubernetesResources(ctx, kubernetesProvider)
	}

	if checkPassed {
		checkPassed = check.ServerConnection(kubernetesProvider)
	}

	if checkPassed {
		log.Info().Msg(fmt.Sprintf(utils.Green, "All checks are passed."))
	} else {
		log.Error().
			Str("command1", fmt.Sprintf("kubeshark %s", cleanCmd.Use)).
			Str("command2", fmt.Sprintf("kubeshark %s", deployCmd.Use)).
			Msg(fmt.Sprintf(utils.Red, "There are issues in your deployment! Run these commands:"))
		os.Exit(1)
	}
}
