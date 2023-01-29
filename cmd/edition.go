package cmd

import (
	"fmt"
	"strings"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/spf13/cobra"
)

var editionCmd = &cobra.Command{
	Use:   "edition",
	Short: fmt.Sprintf("Print the current edition of %s.", misc.Software),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(strings.Title(config.Config.Edition))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editionCmd)
}
