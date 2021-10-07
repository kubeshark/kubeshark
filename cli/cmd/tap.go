package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/telemetry"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/uiUtils"
)

const analysisMessageToConfirm = `NOTE: running mizu with --analysis flag will upload recorded traffic for further analysis and enriched presentation options.`

var tapCmd = &cobra.Command{
	Use:   "tap [POD REGEX]",
	Short: "Record ingoing traffic of a kubernetes pod",
	Long: `Record the ingoing traffic of a kubernetes pod.
Supported protocols are HTTP and gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("tap", config.Config.Tap)
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

		if config.Config.Tap.Analysis {
			logger.Log.Infof(analysisMessageToConfirm)
			if !uiUtils.AskForConfirmation("Would you like to proceed [Y/n]: ") {
				logger.Log.Infof("You can always run mizu without analysis, aborting")
				os.Exit(0)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tapCmd)

	defaultTapConfig := configStructs.TapConfig{}
	defaults.Set(&defaultTapConfig)

	tapCmd.Flags().Uint16P(configStructs.GuiPortTapName, "p", defaultTapConfig.GuiPort, "Provide a custom port for the web interface webserver")
	tapCmd.Flags().StringSliceP(configStructs.NamespacesTapName, "n", defaultTapConfig.Namespaces, "Namespaces selector")
	tapCmd.Flags().Bool(configStructs.AnalysisTapName, defaultTapConfig.Analysis, "Uploads traffic to UP9 for further analysis (Beta)")
	tapCmd.Flags().BoolP(configStructs.AllNamespacesTapName, "A", defaultTapConfig.AllNamespaces, "Tap all namespaces")
	tapCmd.Flags().StringSliceP(configStructs.PlainTextFilterRegexesTapName, "r", defaultTapConfig.PlainTextFilterRegexes, "List of regex expressions that are used to filter matching values from text/plain http bodies")
	tapCmd.Flags().Bool(configStructs.DisableRedactionTapName, defaultTapConfig.DisableRedaction, "Disables redaction of potentially sensitive request/response headers and body values")
	tapCmd.Flags().String(configStructs.HumanMaxEntriesDBSizeTapName, defaultTapConfig.HumanMaxEntriesDBSize, "Override the default max entries db size")
	tapCmd.Flags().Bool(configStructs.DryRunTapName, defaultTapConfig.DryRun, "Preview of all pods matching the regex, without tapping them")
	tapCmd.Flags().String(configStructs.EnforcePolicyFile, defaultTapConfig.EnforcePolicyFile, "Yaml file path with policy rules")
	tapCmd.Flags().String(configStructs.ContractFile, defaultTapConfig.ContractFile, "OAS/Swagger file to validate to monitor the contracts")

	tapCmd.Flags().String(configStructs.EnforcePolicyFileDeprecated, defaultTapConfig.EnforcePolicyFileDeprecated, "Yaml file with policy rules")
	tapCmd.Flags().MarkDeprecated(configStructs.EnforcePolicyFileDeprecated, fmt.Sprintf("Use --%s instead", configStructs.EnforcePolicyFile))
}
