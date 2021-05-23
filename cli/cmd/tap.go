package cmd

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/mizu"
)

var tapCmd = &cobra.Command{
	Use:   "tap [POD REGEX]",
	Short: "Record ingoing traffic of a kubernetes pod",
	Long: `Record the ingoing traffic of a kubernetes pod.
 Supported protocols are HTTP and gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("POD REGEX argument is required")
		} else if len(args) > 1 {
			return errors.New("Unexpected number of arguments")
		}

		regex, err := regexp.Compile(args[0])
		if err != nil {

			return errors.New(fmt.Sprintf("%s is not a valid regex %s", args[0], err))
			return nil
		}

		mizu.Run(regex)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tapCmd)

	tapCmd.Flags().Uint16VarP(&config.Configuration.GuiPort, "gui-port", "p", 8899, "Provide a custom port for the web interface webserver")
	tapCmd.Flags().StringVarP(&config.Configuration.Namespace, "namespace", "n", "", "Namespace selector")
	tapCmd.Flags().StringVarP(&config.Configuration.KubeConfigPath, "kubeconfig", "k", "", "Path to kubeconfig file")
	tapCmd.Flags().StringVarP(&config.Configuration.MizuImage, "mizu-image", "", "gcr.io/up9-docker-hub/mizu/develop:latest", "Custom image for mizu collector")
	tapCmd.Flags().Uint16VarP(&config.Configuration.MizuPodPort, "mizu-port", "", 8899, "Port which mizu cli will attempt to forward from the mizu collector pod")
}
