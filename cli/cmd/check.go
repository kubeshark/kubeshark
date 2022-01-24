package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/telemetry"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the Mizu installation for potential problems",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("check", nil)
		runMizuCheck()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
