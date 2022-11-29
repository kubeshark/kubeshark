package cmd

import (
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the web UI in the browser.",
	RunE: func(cmd *cobra.Command, args []string) error {
		runOpen()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
