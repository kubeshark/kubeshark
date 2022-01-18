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
	viewCmd.Flags().StringP(configStructs.UrlViewName, "u", defaultViewConfig.Url, "Provide a custom host")
	viewCmd.Flags().StringP(configStructs.ProxyTypeViewName, "t", defaultViewConfig.ProxyType, "Provide a custom proxy type")

	viewCmd.Flags().MarkHidden(configStructs.UrlViewName)
}
