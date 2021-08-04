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
	Short: "Generate config with default values",
	RunE: func(cmd *cobra.Command, args []string) error {
		template, err := mizu.GetConfigWithDefaults()
		if err != nil {
			mizu.Log.Errorf("Failed generating config with defaults %v", err)
			return nil
		}
		if regenerateFile {
			data := []byte(template)
			if err := ioutil.WriteFile(mizu.GetConfigFilePath(), data, 0644); err != nil {
				mizu.Log.Errorf("Failed writing config %v", err)
				return nil
			}
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
	configCmd.Flags().BoolVarP(&regenerateFile, "regenerate", "r", false, fmt.Sprintf("Regenerate the config file with default values %s", mizu.GetConfigFilePath()))
}
