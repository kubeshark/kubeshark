package cmd

import (
	"fmt"
	"strings"

	"github.com/kubeshark/kubeshark/cli/bucket"
	"github.com/kubeshark/kubeshark/cli/config"
	"github.com/kubeshark/kubeshark/logger"
)

func runKubesharkInstall() {
	// TODO: Remove this function
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
	sb.WriteString("Hello! This command can be used to install Kubeshark Pro edition on your Kubernetes cluster.")
	sb.WriteString("\nPlease run:")
	sb.WriteString("\n\tkubeshark install -o | kubectl apply -n kubeshark -f -")
	sb.WriteString("\n\nor use helm chart as described in https://getkubeshark.io/docs/installing-kubeshark/centralized-installation\n")

	fmt.Print(sb.String())
}
