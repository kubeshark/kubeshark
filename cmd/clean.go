package cmd

import (
	"fmt"

	"github.com/kubeshark/kubeshark/misc"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: fmt.Sprintf("Removes all %s resources", misc.Software),
	RunE: func(cmd *cobra.Command, args []string) error {
		performCleanCommand()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
