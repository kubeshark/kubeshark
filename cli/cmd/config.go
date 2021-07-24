package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"os"
)

var outputFileName string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate example config file to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		template := config.GetTemplateConfig()
		if outputFileName != "" {
			data := []byte(template)
			_ = os.WriteFile(outputFileName, data, 0644)
			fmt.Println(fmt.Sprintf("Template File written to %s", outputFileName))
		} else {
			fmt.Println(template)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVarP(&outputFileName, "file", "f", "", "Save content to local file")
}
