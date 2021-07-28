package cmd

import (
	"github.com/spf13/cobra"
)

type MizuDemoOptions struct {
	GuiPort            uint16
	Analyze            bool
	AnalyzeDestination string
}

var mizuDemoOptions = &MizuDemoOptions{}

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Record ingoing traffic of a kubernetes pod",
	Long: `Record the ingoing traffic of a kubernetes pod.
Supported protocols are HTTP and gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		RunMizuTapDemo(mizuDemoOptions)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)

	demoCmd.Flags().Uint16VarP(&mizuDemoOptions.GuiPort, "gui-port", "p", 8899, "Provide a custom port for the web interface webserver")
	demoCmd.Flags().BoolVar(&mizuDemoOptions.Analyze, "analyze", false, "Uploads traffic to UP9 cloud for further analysis (Beta)")
	demoCmd.Flags().StringVar(&mizuDemoOptions.AnalyzeDestination, "dest", "up9.app", "Destination environment")
}
