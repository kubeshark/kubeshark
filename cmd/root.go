package cmd

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubeshark",
	Short: fmt.Sprintf("%s: %s", misc.Software, misc.Description),
	Long: fmt.Sprintf(`%s: %s
An extensible Kubernetes-aware network sniffer and kernel tracer.
For more info: %s`, misc.Software, misc.Description, misc.Website),
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
	rootCmd.PersistentFlags().BoolP(config.DebugFlag, "d", false, "Enable debug mode")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the tapCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
