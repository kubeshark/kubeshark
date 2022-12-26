package cmd

import (
	"errors"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var tapCmd = &cobra.Command{
	Use:   "tap [POD REGEX]",
	Short: "Record and see the network traffic in your Kubernetes cluster.",
	Long:  "Record and see the network traffic in your Kubernetes cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		tap()
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			config.Config.Tap.PodRegexStr = args[0]
		} else if len(args) > 1 {
			return errors.New("unexpected number of arguments")
		}

		if err := config.Config.Tap.Validate(); err != nil {
			return errormessage.FormatError(err)
		}

		log.Info().
			Str("limit", config.Config.Tap.HumanMaxEntriesDBSize).
			Msg("Kubeshark will store the traffic up to a limit. Oldest entries will be removed once the limit is reached.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tapCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	tapCmd.Flags().StringP(configStructs.TagLabel, "t", defaultTapConfig.Tag, "The tag of the Docker images that are going to be pulled.")
	tapCmd.Flags().Uint16P(configStructs.ProxyPortLabel, "p", defaultTapConfig.ProxyPort, "Provide a custom port for the web interface webserver.")
	tapCmd.Flags().StringSliceP(configStructs.NamespacesLabel, "n", defaultTapConfig.Namespaces, "Namespaces selector.")
	tapCmd.Flags().BoolP(configStructs.AllNamespacesLabel, "A", defaultTapConfig.AllNamespaces, "Tap all namespaces.")
	tapCmd.Flags().Bool(configStructs.EnableRedactionLabel, defaultTapConfig.EnableRedaction, "Enables redaction of potentially sensitive request/response headers and body values.")
	tapCmd.Flags().String(configStructs.HumanMaxEntriesDBSizeLabel, defaultTapConfig.HumanMaxEntriesDBSize, "Override the default max entries db size.")
	tapCmd.Flags().String(configStructs.InsertionFilterName, defaultTapConfig.InsertionFilter, "Set the insertion filter. Accepts string or a file path.")
	tapCmd.Flags().Bool(configStructs.DryRunLabel, defaultTapConfig.DryRun, "Preview of all pods matching the regex, without tapping them.")
	tapCmd.Flags().Bool(configStructs.ServiceMeshName, defaultTapConfig.ServiceMesh, "Record decrypted traffic if the cluster is configured with a service mesh and with mtls.")
	tapCmd.Flags().Bool(configStructs.TlsName, defaultTapConfig.Tls, "Record tls traffic.")
	tapCmd.Flags().Bool(configStructs.ProfilerName, defaultTapConfig.Profiler, "Run pprof server.")
}
