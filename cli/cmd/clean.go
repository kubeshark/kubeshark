package cmd

import (
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes all mizu resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		performCleanCommand()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
