package cmd

import (
	"context"
	"embed"
	"fmt"

	"github.com/up9inc/mizu/cli/cmd/check"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/logger"
)

var (
	//go:embed permissionFiles
	embedFS embed.FS
)

func runMizuCheck() {
	logger.Log.Infof("Mizu checks\n===================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel will be called when this function exits

	kubernetesProvider, kubernetesVersion, checkPassed := check.KubernetesApi()

	if checkPassed {
		checkPassed = check.KubernetesVersion(kubernetesVersion)
	}

	if config.Config.Check.PreTap || config.Config.Check.PreInstall || config.Config.Check.ImagePull {
		if config.Config.Check.PreTap {
			if checkPassed {
				checkPassed = check.TapKubernetesPermissions(ctx, embedFS, kubernetesProvider)
			}
		} else if config.Config.Check.PreInstall {
			if checkPassed {
				checkPassed = check.InstallKubernetesPermissions(ctx, kubernetesProvider)
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
		logger.Log.Infof("\nStatus check results are %v", fmt.Sprintf(uiUtils.Green, "√"))
	} else {
		logger.Log.Errorf("\nStatus check results are %v", fmt.Sprintf(uiUtils.Red, "✗"))
	}
}
