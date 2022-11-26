package cmd

import (
	"fmt"
	"log"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate config with default values",
	RunE: func(cmd *cobra.Command, args []string) error {
		configWithDefaults, err := config.GetConfigWithDefaults()
		if err != nil {
			log.Printf("Failed generating config with defaults, err: %v", err)
			return nil
		}

		if config.Config.Config.Regenerate {
			if err := config.WriteConfig(configWithDefaults); err != nil {
				log.Printf("Failed writing config with defaults, err: %v", err)
				return nil
			}

			log.Printf("Template File written to %s", fmt.Sprintf(utils.Purple, config.Config.ConfigFilePath))
		} else {
			template, err := utils.PrettyYaml(configWithDefaults)
			if err != nil {
				log.Printf("Failed converting config with defaults to yaml, err: %v", err)
				return nil
			}

			log.Printf("Writing template config.\n%v", template)
			fmt.Printf("%v", template)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	defaultConfig := config.CreateDefaultConfig()
	if err := defaults.Set(&defaultConfig); err != nil {
		log.Print(err)
	}

	configCmd.Flags().BoolP(configStructs.RegenerateConfigName, "r", defaultConfig.Config.Regenerate, fmt.Sprintf("Regenerate the config file with default values to path %s or to chosen path using --%s", defaultConfig.ConfigFilePath, config.ConfigFilePathCommandName))
}
