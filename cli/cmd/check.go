package cmd

import (
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/telemetry"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks that mizu was installed successfully",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("check", nil)
		runMizuCheck()
		return nil
	},
}

func init() {
	installCmd.AddCommand(checkCmd)

	defaultCheckConfig := configStructs.CheckConfig{}
	defaults.Set(&defaultCheckConfig)

	checkCmd.Flags().StringP(configStructs.ServerUrlCheckName, "u", defaultCheckConfig.ServerUrl, "Provide a custom server url")
}
