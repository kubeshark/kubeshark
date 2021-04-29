package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/mizu"
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{}
)
func init() {
	rootCmd.Use = "cmd pod-query"
	rootCmd.Short = "Tail HTTP traffic from multiple pods"
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return rootCmd.Help()
		}

		mizu.Run()
		return nil
	}

	rootCmd.Flags().BoolVarP(&config.Configuration.DisplayVersion, "version", "v", false, "Print the version and exit")
	rootCmd.Flags().BoolVarP(&config.Configuration.Quiet, "quiet", "q", false, "No stdout output")
	rootCmd.Flags().BoolVarP(&config.Configuration.NoDashboard, "no-dashboard", "", false, "Dont host a dashboard")
	rootCmd.Flags().Uint16VarP(&config.Configuration.DashboardPort, "dashboard-port", "p", 8899, "Provide a custom port for the dashboard webserver")
	rootCmd.Flags().StringVarP(&config.Configuration.Namespace, "namespace", "n", "", "Namespace selector")
	rootCmd.Flags().BoolVarP(&config.Configuration.AllNamespaces, "all-namespaces", "A", false, "Select all namespaces")
	rootCmd.Flags().StringVarP(&config.Configuration.KubeConfigPath, "kubeconfig", "k", "", "Path to kubeconfig file")
	rootCmd.Flags().StringVarP(&config.Configuration.MizuImage, "mizu-image", "", "gcr.io/up9-docker-hub/mizu/develop:latest", "Custom image for mizu collector")
	rootCmd.Flags().Uint16VarP(&config.Configuration.MizuPodPort, "mizu-port", "", 8899, "Port which mizu cli will attempt to forward from the mizu collector pod")
	rootCmd.Flags().StringVarP(&config.Configuration.TappedPodName, "pod", "", "", "View traffic of this pod")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
