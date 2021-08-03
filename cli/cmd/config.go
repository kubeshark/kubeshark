package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/uiUtils"
	"io/ioutil"
)

var regenerateFile bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate example config file to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		template := mizu.GetTemplateConfig()
		if regenerateFile {
			data := []byte(template)
			_ = ioutil.WriteFile(mizu.GetConfigFilePath(), data, 0644)
			mizu.Log.Infof(fmt.Sprintf("Template File written to %s", fmt.Sprintf(uiUtils.Purple, mizu.GetConfigFilePath())))
		} else {
			mizu.Log.Debugf("Writing template config.\n%v", template)
			fmt.Printf("%v", template)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().BoolVarP(&regenerateFile, "file", "f", false, "Save content to local file")
}
