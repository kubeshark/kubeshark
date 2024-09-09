package cmd

import (
	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var pprofCmd = &cobra.Command{
	Use:   "pprof",
	Short: "Select a Kubeshark container and open the pprof web UI in the browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		runPprof()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pprofCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	pprofCmd.Flags().Uint16(configStructs.ProxyFrontPortLabel, defaultTapConfig.Proxy.Front.Port, "Provide a custom port for the proxy/port-forward")
	pprofCmd.Flags().String(configStructs.ProxyHostLabel, defaultTapConfig.Proxy.Host, "Provide a custom host for the proxy/port-forward")
	pprofCmd.Flags().StringP(configStructs.ReleaseNamespaceLabel, "s", defaultTapConfig.Release.Namespace, "Release namespace of Kubeshark")
	pprofCmd.Flags().Uint16(configStructs.PprofPortLabel, defaultTapConfig.Pprof.Port, "Provide a custom port for the pprof server")
	pprofCmd.Flags().String(configStructs.PprofViewLabel, defaultTapConfig.Pprof.View, "Change the default view of the pprof web interface")
}
