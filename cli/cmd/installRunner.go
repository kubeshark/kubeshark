package cmd

import (
	"fmt"
	"strings"

	"github.com/up9inc/mizu/cli/bucket"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/logger"
)

func runMizuInstall() {
	if config.Config.Install.Out {
		bucketProvider := bucket.NewProvider(config.Config.Install.TemplateUrl, bucket.DefaultTimeout)
		installTemplate, err := bucketProvider.GetInstallTemplate(config.Config.Install.TemplateName)
		if err != nil {
			logger.Log.Errorf("Failed getting install template, err: %v", err)
			return
		}

		fmt.Print(installTemplate)
		return
	}

	var sb strings.Builder
	sb.WriteString("Hello! This command can be used to install Mizu Pro edition on your Kubernetes cluster.")
	sb.WriteString("\nPlease run:")
	sb.WriteString("\n\tmizu install -o | kubectl apply -f -")
	sb.WriteString("\n\nor use helm chart as described in https://getmizu.io/docs/installing-mizu/centralized-installation\n")

	fmt.Print(sb.String())
}
