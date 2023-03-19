package cmd

import (
	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Open the web UI (front-end) in the browser via proxy/port-forward",
	RunE: func(cmd *cobra.Command, args []string) error {
		runProxy(true, false)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	proxyCmd.Flags().Uint16(configStructs.ProxyFrontPortLabel, defaultTapConfig.Proxy.Front.SrcPort, "Provide a custom port for the front-end proxy/port-forward")
	proxyCmd.Flags().Uint16(configStructs.ProxyHubPortLabel, defaultTapConfig.Proxy.Hub.SrcPort, "Provide a custom port for the Hub proxy/port-forward")
	proxyCmd.Flags().String(configStructs.ProxyHostLabel, defaultTapConfig.Proxy.Host, "Provide a custom host for the proxy/port-forward")
}
