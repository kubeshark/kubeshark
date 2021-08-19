package cmd

import (
	"fmt"
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/cli/uiUtils"
	"io/ioutil"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate config with default values",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("config", config.Config.Config)

		template, err := config.GetConfigWithDefaults()
		if err != nil {
			logger.Log.Errorf("Failed generating config with defaults %v", err)
			return nil
		}
		if config.Config.Config.Regenerate {
			data := []byte(template)
			if err := ioutil.WriteFile(config.GetConfigFilePath(), data, 0644); err != nil {
				logger.Log.Errorf("Failed writing config %v", err)
				return nil
			}
			logger.Log.Infof(fmt.Sprintf("Template File written to %s", fmt.Sprintf(uiUtils.Purple, config.GetConfigFilePath())))
		} else {
			logger.Log.Debugf("Writing template config.\n%v", template)
			fmt.Printf("%v", template)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	defaultConfigConfig := configStructs.ConfigConfig{}
	defaults.Set(&defaultConfigConfig)

	configCmd.Flags().BoolP(configStructs.RegenerateConfigName, "r", defaultConfigConfig.Regenerate, fmt.Sprintf("Regenerate the config file with default values %s", config.GetConfigFilePath()))
}
