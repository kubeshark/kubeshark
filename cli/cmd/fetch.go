package cmd

import (
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download recorded traffic to files",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Flags().Visit(mizu.InitFlag)

		go mizu.ReportRun("fetch", mizu.Config.Fetch)
		if isCompatible, err := mizu.CheckVersionCompatibility(mizu.Config.Fetch.MizuPort); err != nil {
			return err
		} else if !isCompatible {
			return nil
		}
		RunMizuFetch()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	defaultFetchConfig := mizu.FetchConfig{}
	defaults.Set(&defaultFetchConfig)

	fetchCmd.Flags().StringP("directory", "d", defaultFetchConfig.Directory, "Provide a custom directory for fetched entries")
	fetchCmd.Flags().Int("from", defaultFetchConfig.FromTimestamp, "Custom start timestamp for fetched entries")
	fetchCmd.Flags().Int("to", defaultFetchConfig.ToTimestamp, "Custom end timestamp fetched entries")
	fetchCmd.Flags().Uint16P("port", "p", defaultFetchConfig.MizuPort, "Custom port for mizu")
}
