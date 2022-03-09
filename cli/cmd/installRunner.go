package cmd

import (
	"fmt"
	"github.com/up9inc/mizu/cli/bucket"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/shared/logger"
)

func runMizuInstall() {
	bucketProvider := bucket.NewProvider(config.Config.Install.TemplateUrl, bucket.DefaultTimeout)
	installTemplate, err := bucketProvider.GetInstallTemplate(config.Config.Install.TemplateName)
	if err != nil {
		logger.Log.Errorf("Failed getting install template, err: %v", err)
		return
	}

	fmt.Print(installTemplate)
}
