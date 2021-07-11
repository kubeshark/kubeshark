package cmd

import (
	"github.com/spf13/cobra"
)

type MizuViewOptions struct {
	GuiPort                uint16
}

var mizuViewOptions = &MizuViewOptions{}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Open GUI in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		runMizuView(mizuViewOptions)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)

	viewCmd.Flags().Uint16VarP(&mizuViewOptions.GuiPort, "gui-port", "p", 8899, "Provide a custom port for the web interface webserver")

}
