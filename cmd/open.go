package cmd

import (
	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the web UI (front-end) in the browser via proxy/port-forward.",
	RunE: func(cmd *cobra.Command, args []string) error {
		runOpen()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	openCmd.Flags().Uint16(configStructs.ProxyPortFrontLabel, defaultTapConfig.Front.SrcPort, "Provide a custom port for the front-end proxy/port-forward.")
	openCmd.Flags().Uint16(configStructs.ProxyPortHubLabel, defaultTapConfig.Hub.SrcPort, "Provide a custom port for the Hub proxy/port-forward.")
	openCmd.Flags().String(configStructs.ProxyHostLabel, defaultTapConfig.ProxyHost, "Provide a custom host for the proxy/port-forward.")
}
