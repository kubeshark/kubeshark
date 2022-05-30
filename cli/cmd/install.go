package cmd

import (
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/logger"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs mizu components",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("install", nil)
		runMizuInstall()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	defaultInstallConfig := configStructs.InstallConfig{}
	if err := defaults.Set(&defaultInstallConfig); err != nil {
		logger.Log.Debug(err)
	}

	installCmd.Flags().BoolP(configStructs.OutInstallName, "o", defaultInstallConfig.Out, "print (to stdout) Kubernetes manifest used to install Mizu Pro edition")
}
