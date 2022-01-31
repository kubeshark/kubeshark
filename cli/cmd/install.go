package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if config.Config.IsNsRestrictedMode() {
			return fmt.Errorf("install is not supported in restricted namespace mode")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

