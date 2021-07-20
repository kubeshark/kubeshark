package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
)

type MizuFetchOptions struct {
	FromTimestamp int64
	ToTimestamp   int64
	Directory     string
	MizuPort      uint16
}

var mizuFetchOptions = MizuFetchOptions{}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download recorded traffic to files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isCompatible, err := mizu.CheckVersionCompatibility(mizuFetchOptions.MizuPort); err != nil {
			return err
		} else if !isCompatible {
			return nil
		}
		RunMizuFetch(&mizuFetchOptions)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().StringVarP(&mizuFetchOptions.Directory, "directory", "d", ".", "Provide a custom directory for fetched entries")
	fetchCmd.Flags().Int64Var(&mizuFetchOptions.FromTimestamp, "from", 0, "Custom start timestamp for fetched entries")
	fetchCmd.Flags().Int64Var(&mizuFetchOptions.ToTimestamp, "to", 0, "Custom end timestamp fetched entries")
	fetchCmd.Flags().Uint16VarP(&mizuFetchOptions.MizuPort, "port", "p", 8899, "Custom port for mizu")
}
