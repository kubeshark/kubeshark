package cmd

import (
	"github.com/spf13/cobra"
)

var pprofCmd = &cobra.Command{
	Use:   "pprof",
	Short: "Select a Kubeshark container and open the pprof web UI in the browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		runPprof()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pprofCmd)
}
