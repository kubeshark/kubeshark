package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/shared/logger"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs mizu components",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("install", nil)
		logger.Log.Infof("This command has been deprecated, please use helm as described below.\n\n")

		logger.Log.Infof("To install stable build of Mizu on your cluster using helm, run the following command:")
		logger.Log.Infof("    helm install mizu https://static.up9.com/mizu/helm --namespace=mizu-ent --create-namespace\n\n")

		logger.Log.Infof("To install development build of Mizu on your cluster using helm, run the following command:")
		logger.Log.Infof("    helm install mizu https://static.up9.com/mizu/helm-develop --namespace=mizu-ent --create-namespace")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
