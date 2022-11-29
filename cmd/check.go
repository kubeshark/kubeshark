package cmd

import (
	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the Kubeshark installation for potential problems",
	RunE: func(cmd *cobra.Command, args []string) error {
		runKubesharkCheck()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	defaultCheckConfig := configStructs.CheckConfig{}
	if err := defaults.Set(&defaultCheckConfig); err != nil {
		log.Error().Err(err).Send()
	}

	checkCmd.Flags().Bool(configStructs.PreTapCheckName, defaultCheckConfig.PreTap, "Check pre-tap Kubeshark installation for potential problems")
	checkCmd.Flags().Bool(configStructs.ImagePullCheckName, defaultCheckConfig.ImagePull, "Test connectivity to container image registry by creating and removing a temporary pod in 'default' namespace")
}
