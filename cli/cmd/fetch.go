package cmd

import (
	"github.com/spf13/cobra"
)

type MizuFetchOptions struct {
	Limit        uint16
	Directory	 string
}

var mizuFetchOptions = MizuFetchOptions{}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download recorded traffic to files",
	RunE: func(cmd *cobra.Command, args []string) error {
		RunMizuFetch(&mizuFetchOptions)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().Uint16VarP(&mizuFetchOptions.Limit, "limit", "l", 1000, "Provide a custom limit for entries to fetch")
	fetchCmd.Flags().StringVarP(&mizuFetchOptions.Directory, "directory", "d", ".", "Provide a custom directory for fetched entries")
}
