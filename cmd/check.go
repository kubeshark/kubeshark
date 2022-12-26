package cmd

import (
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the Kubeshark resources for potential problems",
	RunE: func(cmd *cobra.Command, args []string) error {
		runKubesharkCheck()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
