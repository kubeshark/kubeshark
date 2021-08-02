package cmd

import (
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Open GUI in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Flags().Visit(mizu.InitFlag)

		go mizu.ReportRun("view", mizu.Config.View)
		if isCompatible, err := mizu.CheckVersionCompatibility(mizu.Config.View.GuiPort); err != nil {
			return err
		} else if !isCompatible {
			return nil
		}
		runMizuView()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)

	defaultViewConfig := mizu.ViewConfig{}
	defaults.Set(&defaultViewConfig)

	viewCmd.Flags().Uint16P("gui-port", "p", defaultViewConfig.GuiPort, "Provide a custom port for the web interface webserver")
	viewCmd.Flags().StringP("kube-config", "k", defaultViewConfig.KubeConfigPath, "Path to kube-config file")
}
