package cmd

import (
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/apiserver"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu/version"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/cli/uiUtils"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download recorded traffic to files",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("fetch", config.Config.Fetch)

		if err := apiserver.Provider.Init(GetApiServerUrl(), 1); err != nil {
			logger.Log.Errorf(uiUtils.Error, "Couldn't connect to API server, make sure one running")
			return nil
		}

		if isCompatible, err := version.CheckVersionCompatibility(); err != nil {
			return err
		} else if !isCompatible {
			return nil
		}
		RunMizuFetch()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	defaultFetchConfig := configStructs.FetchConfig{}
	defaults.Set(&defaultFetchConfig)

	fetchCmd.Flags().StringP(configStructs.DirectoryFetchName, "d", defaultFetchConfig.Directory, "Provide a custom directory for fetched entries")
	fetchCmd.Flags().Int(configStructs.FromTimestampFetchName, defaultFetchConfig.FromTimestamp, "Custom start timestamp for fetched entries")
	fetchCmd.Flags().Int(configStructs.ToTimestampFetchName, defaultFetchConfig.ToTimestamp, "Custom end timestamp fetched entries")
	fetchCmd.Flags().Uint16P(configStructs.GuiPortFetchName, "p", defaultFetchConfig.GuiPort, "Provide a custom port for the web interface webserver")
}
