package cmd

import (
	"fmt"
	"path"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: fmt.Sprintf("Generate %s config with default values", misc.Software),
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.Config.Config.Regenerate {
			defaultConfig := config.CreateDefaultConfig()
			if err := defaults.Set(&defaultConfig); err != nil {
				log.Error().Err(err).Send()
				return nil
			}
			if err := config.WriteConfig(&defaultConfig); err != nil {
				log.Error().Err(err).Msg("Failed generating config with defaults.")
				return nil
			}

			log.Info().Str("config-path", config.ConfigFilePath).Msg("Template file written to config path.")
		} else {
			template, err := utils.PrettyYaml(config.Config)
			if err != nil {
				log.Error().Err(err).Msg("Failed converting config with defaults to YAML.")
				return nil
			}

			log.Debug().Str("template", template).Msg("Printing template config...")
			fmt.Printf("%v", template)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	defaultConfig := config.CreateDefaultConfig()
	if err := defaults.Set(&defaultConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	configCmd.Flags().BoolP(configStructs.RegenerateConfigName, "r", defaultConfig.Config.Regenerate, fmt.Sprintf("Regenerate the config file with default values to path %s", path.Join(misc.GetDotFolderPath(), "config.yaml")))
}
