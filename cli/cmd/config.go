package cmd

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/cli/config"
	"github.com/kubeshark/kubeshark/cli/config/configStructs"
	"github.com/kubeshark/kubeshark/cli/uiUtils"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate config with default values",
	RunE: func(cmd *cobra.Command, args []string) error {
		configWithDefaults, err := config.GetConfigWithDefaults()
		if err != nil {
			logger.Log.Errorf("Failed generating config with defaults, err: %v", err)
			return nil
		}

		if config.Config.Config.Regenerate {
			if err := config.WriteConfig(configWithDefaults); err != nil {
				logger.Log.Errorf("Failed writing config with defaults, err: %v", err)
				return nil
			}

			logger.Log.Infof(fmt.Sprintf("Template File written to %s", fmt.Sprintf(uiUtils.Purple, config.Config.ConfigFilePath)))
		} else {
			template, err := uiUtils.PrettyYaml(configWithDefaults)
			if err != nil {
				logger.Log.Errorf("Failed converting config with defaults to yaml, err: %v", err)
				return nil
			}

			logger.Log.Debugf("Writing template config.\n%v", template)
			fmt.Printf("%v", template)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	defaultConfig := config.ConfigStruct{}
	if err := defaults.Set(&defaultConfig); err != nil {
		logger.Log.Debug(err)
	}

	configCmd.Flags().BoolP(configStructs.RegenerateConfigName, "r", defaultConfig.Config.Regenerate, fmt.Sprintf("Regenerate the config file with default values to path %s or to chosen path using --%s", defaultConfig.ConfigFilePath, config.ConfigFilePathCommandName))
}
