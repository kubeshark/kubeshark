package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download recorded traffic",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Not implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
