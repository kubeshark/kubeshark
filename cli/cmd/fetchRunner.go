package cmd

import (
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/uiUtils"
)

func RunMizuFetch() {
	if err := apiserver.Provider.Init(GetApiServerUrl(), 5); err != nil {
		logger.Log.Errorf(uiUtils.Error, "Couldn't connect to API server, check logs")
	}

	zipReader, err := apiserver.Provider.GetHars(config.Config.Fetch.FromTimestamp, config.Config.Fetch.ToTimestamp)
	if err != nil {
		logger.Log.Errorf("Failed fetch data from API server %v", err)
		return
	}

	_ = fsUtils.Unzip(zipReader, config.Config.Fetch.Directory)
}
