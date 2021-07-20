package cmd

import (
	"fmt"
	"github.com/up9inc/mizu/cli/mizu"
	"strconv"
	"time"

	"github.com/spf13/cobra"
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
			fmt.Printf("Version: %s \nBranch: %s (%s) \n", mizu.SemVer, mizu.Branch, mizu.GitCommitHash)
			fmt.Printf("Build Time: %s (%s)\n", mizu.BuildTimestamp, time.Unix(timeStampInt, 0))

		} else {
			fmt.Printf("Version: %s (%s)\n", mizu.SemVer, mizu.Branch)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVarP(&mizuVersionOptions.DebugInfo, "debug", "d", false, "Provide all information about version")

}
