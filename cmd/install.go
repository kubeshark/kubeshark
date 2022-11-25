package cmd

import (
	"log"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs kubeshark components",
	RunE: func(cmd *cobra.Command, args []string) error {
		runKubesharkInstall()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	defaultInstallConfig := configStructs.InstallConfig{}
	if err := defaults.Set(&defaultInstallConfig); err != nil {
		log.Print(err)
	}

	installCmd.Flags().BoolP(configStructs.OutInstallName, "o", defaultInstallConfig.Out, "print (to stdout) Kubernetes manifest used to install Kubeshark Pro edition")
}
