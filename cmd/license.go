package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubeshark/kubeshark/config"
)

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Print the license loaded string",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(config.Config.License)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(licenseCmd)
}
