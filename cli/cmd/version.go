package cmd

import (
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/configStructs"
	"strconv"
	"time"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Flags().Visit(mizu.InitFlag)

		go mizu.ReportRun("version", mizu.Config.Version)
		if mizu.Config.Version.DebugInfo {
			timeStampInt, _ := strconv.ParseInt(mizu.BuildTimestamp, 10, 0)
			mizu.Log.Infof("Version: %s \nBranch: %s (%s)", mizu.SemVer, mizu.Branch, mizu.GitCommitHash)
			mizu.Log.Infof("Build Time: %s (%s)", mizu.BuildTimestamp, time.Unix(timeStampInt, 0))

		} else {
			mizu.Log.Infof("Version: %s (%s)", mizu.SemVer, mizu.Branch)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	defaultVersionConfig := configStructs.VersionConfig{}
	defaults.Set(&defaultVersionConfig)

	versionCmd.Flags().BoolP(configStructs.DebugInfoVersionName, "d", defaultVersionConfig.DebugInfo, "Provide all information about version")

}
