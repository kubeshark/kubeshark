package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
	"regexp"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{}
func init() {
	rootCmd.Use = "cmd pod-query"
	rootCmd.Short = "Tail HTTP traffic from multiple pods"
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return rootCmd.Help()
		}
		regex := regexp.MustCompile(args[0]) // MustCompile panics if expression cant be compiled into regex
		mizu.Run(regex)
		return nil
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print the version and exit")
	rootCmd.Flags().BoolP("quiet", "q", false, "No stdout output")
	rootCmd.Flags().BoolP("no-dashboard", "", false, "Dont host a dashboard")
	rootCmd.Flags().Uint16P("dashboard-port", "p", 3000, "Provide a custom port for the dashboard webserver")
	rootCmd.Flags().StringP("namespace", "n", "", "Namespace selector")
	rootCmd.Flags().BoolP("all-namespaces", "A", false, "Select all namespaces")
	rootCmd.Flags().StringP("kubeconfig", "k", "", "Path to kubeconfig file")
}
