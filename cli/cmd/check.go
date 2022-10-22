package cmd

import (
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/kubeshark/cli/config/configStructs"
	"github.com/up9inc/kubeshark/logger"
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
		logger.Log.Debug(err)
	}

	checkCmd.Flags().Bool(configStructs.PreTapCheckName, defaultCheckConfig.PreTap, "Check pre-tap Kubeshark installation for potential problems")
	checkCmd.Flags().Bool(configStructs.PreInstallCheckName, defaultCheckConfig.PreInstall, "Check pre-install Kubeshark installation for potential problems")
	checkCmd.Flags().Bool(configStructs.ImagePullCheckName, defaultCheckConfig.ImagePull, "Test connectivity to container image registry by creating and removing a temporary pod in 'default' namespace")
}
