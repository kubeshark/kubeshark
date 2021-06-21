package cmd

import (
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Open GUI in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		runMizuView()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
