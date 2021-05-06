package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/mizu"
)

var tapCmd = &cobra.Command{
	Use:   "tap",
	Short: "Tail HTTP traffic from multiple pods",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return errors.New("Unexpected number of arguments")
		}

		mizu.Run()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tapCmd)

	tapCmd.Flags().BoolVarP(&config.Configuration.DisplayVersion, "version", "v", false, "Print the version and exit")
	tapCmd.Flags().BoolVarP(&config.Configuration.Quiet, "quiet", "q", false, "No stdout output")
	tapCmd.Flags().BoolVarP(&config.Configuration.NoDashboard, "no-dashboard", "", false, "Dont host a dashboard")
	tapCmd.Flags().Uint16VarP(&config.Configuration.DashboardPort, "dashboard-port", "p", 8899, "Provide a custom port for the dashboard webserver")
	tapCmd.Flags().StringVarP(&config.Configuration.Namespace, "namespace", "n", "", "Namespace selector")
	tapCmd.Flags().BoolVarP(&config.Configuration.AllNamespaces, "all-namespaces", "A", false, "Select all namespaces")
	tapCmd.Flags().StringVarP(&config.Configuration.KubeConfigPath, "kubeconfig", "k", "", "Path to kubeconfig file")
	tapCmd.Flags().StringVarP(&config.Configuration.MizuImage, "mizu-image", "", "gcr.io/up9-docker-hub/mizu/develop:latest", "Custom image for mizu collector")
	tapCmd.Flags().Uint16VarP(&config.Configuration.MizuPodPort, "mizu-port", "", 8899, "Port which mizu cli will attempt to forward from the mizu collector pod")
	tapCmd.Flags().StringVarP(&config.Configuration.TappedPodName, "pod", "", "", "View traffic of this pod")
	err := tapCmd.MarkFlagRequired("pod")
	if err != nil {
		panic(err)
	}
}
