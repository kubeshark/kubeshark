package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/kubeshark/kubeshark/cmd/check"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
)

var (
	//go:embed permissionFiles
	embedFS embed.FS
)

func runCheck() {
	log.Info().Msg(fmt.Sprintf("Checking the %s resources...", misc.Software))

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
			Str("command1", fmt.Sprintf("%s %s", misc.Program, cleanCmd.Use)).
			Str("command2", fmt.Sprintf("%s %s", misc.Program, tapCmd.Use)).
			Msg(fmt.Sprintf(utils.Red, fmt.Sprintf("There are issues in your %s resources! Run these commands:", misc.Software)))
		os.Exit(1)
	}
}
