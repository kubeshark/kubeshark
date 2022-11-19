package cmd

import (
	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/cli/config/configStructs"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Open GUI in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		runKubesharkView()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)

	defaultViewConfig := configStructs.ViewConfig{}
	if err := defaults.Set(&defaultViewConfig); err != nil {
		logger.Log.Debug(err)
	}

	viewCmd.Flags().Uint16P(configStructs.GuiPortViewName, "p", defaultViewConfig.GuiPort, "Provide a custom port for the web interface webserver")
	viewCmd.Flags().StringP(configStructs.UrlViewName, "u", defaultViewConfig.Url, "Provide a custom host")

	if err := viewCmd.Flags().MarkHidden(configStructs.UrlViewName); err != nil {
		logger.Log.Debug(err)
	}
}
