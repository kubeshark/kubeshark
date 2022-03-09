package cmd

import (
	"fmt"
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/shared/logger"
)

func runMizuInstall() {
	apiProvider = apiserver.NewProvider(config.Config.Install.TemplateUrl, apiserver.DefaultRetries, apiserver.DefaultTimeout)
	installTemplate, err := apiProvider.GetInstallTemplate(config.Config.Install.TemplateName)
	if err != nil {
		logger.Log.Errorf("Failed getting install template, err: %v", err)
		return
	}

	fmt.Print(installTemplate)
}
