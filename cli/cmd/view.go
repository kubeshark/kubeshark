package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Open GUI in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("Not implemented")
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
