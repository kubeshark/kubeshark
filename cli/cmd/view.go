package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
)

type MizuViewOptions struct {
	GuiPort uint16
}

var mizuViewOptions = &MizuViewOptions{}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Open GUI in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		go mizu.ReportRun("view", mizuFetchOptions)
		if isCompatible, err := mizu.CheckVersionCompatibility(mizuFetchOptions.MizuPort); err != nil {
			return err
		} else if !isCompatible {
			return nil
		}
		runMizuView(mizuViewOptions)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)

	viewCmd.Flags().Uint16VarP(&mizuViewOptions.GuiPort, "gui-port", "p", 8899, "Provide a custom port for the web interface webserver")

}
