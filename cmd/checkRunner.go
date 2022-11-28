package cmd

import (
	"context"
	"embed"
	"fmt"

	"github.com/kubeshark/kubeshark/cmd/check"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
)

var (
	//go:embed permissionFiles
	embedFS embed.FS
)

func runKubesharkCheck() {
	log.Info().Msg("Kubeshark checks...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	kubernetesProvider, kubernetesVersion, checkPassed := check.KubernetesApi()

	if checkPassed {
		checkPassed = check.KubernetesVersion(kubernetesVersion)
	}

	if config.Config.Check.PreTap || config.Config.Check.ImagePull {
		if config.Config.Check.PreTap {
			if checkPassed {
				checkPassed = check.TapKubernetesPermissions(ctx, embedFS, kubernetesProvider)
			}
		}

		if config.Config.Check.ImagePull {
			if checkPassed {
				checkPassed = check.ImagePullInCluster(ctx, kubernetesProvider)
			}
		}
	} else {
		if checkPassed {
			checkPassed = check.KubernetesResources(ctx, kubernetesProvider)
		}

		if checkPassed {
			checkPassed = check.ServerConnection(kubernetesProvider)
		}
	}

	if checkPassed {
		log.Info().Msg(fmt.Sprintf("Status check results are %v", fmt.Sprintf(utils.Green, "√")))
	} else {
		log.Info().Msg(fmt.Sprintf("Status check results are %v", fmt.Sprintf(utils.Red, "✗")))
	}
}
