package cmd

import (
	"fmt"
	"github.com/up9inc/mizu/cli/mizu"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s (%s) %s\n", mizu.Version, mizu.Branch, mizu.GitCommitHash)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
