package cmd

import (
	"strconv"
	"time"

	"github.com/up9inc/kubeshark/cli/config"
	"github.com/up9inc/kubeshark/cli/config/configStructs"
	"github.com/up9inc/kubeshark/logger"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/kubeshark/cli/kubeshark"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.Config.Version.DebugInfo {
			timeStampInt, _ := strconv.ParseInt(kubeshark.BuildTimestamp, 10, 0)
			logger.Log.Infof("Version: %s \nBranch: %s (%s)", kubeshark.Ver, kubeshark.Branch, kubeshark.GitCommitHash)
			logger.Log.Infof("Build Time: %s (%s)", kubeshark.BuildTimestamp, time.Unix(timeStampInt, 0))

		} else {
			logger.Log.Infof("Version: %s (%s)", kubeshark.Ver, kubeshark.Branch)
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
