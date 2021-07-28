package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
	"strconv"
	"time"
)

type MizuVersionOptions struct {
	DebugInfo bool
}

var mizuVersionOptions = &MizuVersionOptions{}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	RunE: func(cmd *cobra.Command, args []string) error {
		go mizu.ReportRun("version", mizuVersionOptions)
		if mizuVersionOptions.DebugInfo {
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

	versionCmd.Flags().BoolVarP(&mizuVersionOptions.DebugInfo, "debug", "d", false, "Provide all information about version")

}
