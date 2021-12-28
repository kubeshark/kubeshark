package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/telemetry"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs mizu components",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("install", nil)
		runMizuInstall()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

