package cmd

import (
	"strconv"
	"time"

	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/logger"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("version", config.Config.Version)

		if config.Config.Version.DebugInfo {
			timeStampInt, _ := strconv.ParseInt(mizu.BuildTimestamp, 10, 0)
			logger.Log.Infof("Version: %s \nBranch: %s (%s)", mizu.Ver, mizu.Branch, mizu.GitCommitHash)
			logger.Log.Infof("Build Time: %s (%s)", mizu.BuildTimestamp, time.Unix(timeStampInt, 0))

		} else {
			logger.Log.Infof("Version: %s (%s)", mizu.Ver, mizu.Branch)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	defaultVersionConfig := configStructs.VersionConfig{}
	if err := defaults.Set(&defaultVersionConfig); err != nil {
		logger.Log.Debug(err)
	}

	versionCmd.Flags().BoolP(configStructs.DebugInfoVersionName, "d", defaultVersionConfig.DebugInfo, "Provide all information about version")

}
