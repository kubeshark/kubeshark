package cmd

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubeshark/version"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubeshark",
	Short: "A web traffic viewer for kubernetes",
	Long: `A web traffic viewer for kubernetes
Further info is available at https://github.com/kubeshark/kubeshark`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(cmd); err != nil {
			log.Fatal().Err(err).Send()
		}

		return nil
	},
}

func init() {
	defaultConfig := config.CreateDefaultConfig()
	if err := defaults.Set(&defaultConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	rootCmd.PersistentFlags().StringSlice(config.SetCommandName, []string{}, fmt.Sprintf("Override values using --%s", config.SetCommandName))
	rootCmd.PersistentFlags().String(config.ConfigFilePathCommandName, defaultConfig.ConfigFilePath, fmt.Sprintf("Override config file path using --%s", config.ConfigFilePathCommandName))
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the tapCmd.
func Execute() {
	go version.CheckNewerVersion()

	cobra.CheckErr(rootCmd.Execute())
}
