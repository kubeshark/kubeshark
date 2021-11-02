package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/telemetry"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes all mizu resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("clean", nil)
		performCleanCommand()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
