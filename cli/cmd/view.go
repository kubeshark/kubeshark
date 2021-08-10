package cmd

import (
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/telemetry"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Open GUI in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("view", config.Config.View)
		runMizuView()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)

	defaultViewConfig := configStructs.ViewConfig{}
	defaults.Set(&defaultViewConfig)

	viewCmd.Flags().Uint16P(configStructs.GuiPortViewName, "p", defaultViewConfig.GuiPort, "Provide a custom port for the web interface webserver")
	viewCmd.Flags().StringP(configStructs.KubeConfigPathViewName, "k", defaultViewConfig.KubeConfigPath, "Path to kube-config file")
}
