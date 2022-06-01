package cmd

import (
	"errors"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/logger"
)

var tapCmd = &cobra.Command{
	Use:   "tap [POD REGEX]",
	Short: "Record ingoing traffic of a kubernetes pod",
	Long: `Record the ingoing traffic of a kubernetes pod.
Supported protocols are HTTP and gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		RunMizuTap()
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

		logger.Log.Infof("Mizu will store up to %s of traffic, old traffic will be cleared once the limit is reached.", config.Config.Tap.HumanMaxEntriesDBSize)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tapCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		logger.Log.Debug(err)
	}

	tapCmd.Flags().Uint16P(configStructs.GuiPortTapName, "p", defaultTapConfig.GuiPort, "Provide a custom port for the web interface webserver")
	tapCmd.Flags().StringSliceP(configStructs.NamespacesTapName, "n", defaultTapConfig.Namespaces, "Namespaces selector")
	tapCmd.Flags().BoolP(configStructs.AllNamespacesTapName, "A", defaultTapConfig.AllNamespaces, "Tap all namespaces")
	tapCmd.Flags().StringSliceP(configStructs.PlainTextFilterRegexesTapName, "r", defaultTapConfig.PlainTextFilterRegexes, "List of regex expressions that are used to filter matching values from text/plain http bodies")
	tapCmd.Flags().Bool(configStructs.EnableRedactionTapName, defaultTapConfig.EnableRedaction, "Enables redaction of potentially sensitive request/response headers and body values")
	tapCmd.Flags().String(configStructs.HumanMaxEntriesDBSizeTapName, defaultTapConfig.HumanMaxEntriesDBSize, "Override the default max entries db size")
	tapCmd.Flags().String(configStructs.InsertionFilterName, defaultTapConfig.InsertionFilter, "Set the insertion filter. Accepts string or a file path.")
	tapCmd.Flags().Bool(configStructs.DryRunTapName, defaultTapConfig.DryRun, "Preview of all pods matching the regex, without tapping them")
	tapCmd.Flags().String(configStructs.EnforcePolicyFile, defaultTapConfig.EnforcePolicyFile, "Yaml file path with policy rules")
	tapCmd.Flags().String(configStructs.ContractFile, defaultTapConfig.ContractFile, "OAS/Swagger file to validate to monitor the contracts")
	tapCmd.Flags().Bool(configStructs.ServiceMeshName, defaultTapConfig.ServiceMesh, "Record decrypted traffic if the cluster is configured with a service mesh and with mtls")
	tapCmd.Flags().Bool(configStructs.TlsName, defaultTapConfig.Tls, "Record tls traffic")
	tapCmd.Flags().Bool(configStructs.ProfilerName, defaultTapConfig.Profiler, "Run pprof server")
}
