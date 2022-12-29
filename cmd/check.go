package cmd

import (
	"fmt"

	"github.com/kubeshark/kubeshark/misc"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: fmt.Sprintf("Check the %s resources for potential problems", misc.Software),
	RunE: func(cmd *cobra.Command, args []string) error {
		runCheck()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
