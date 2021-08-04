package cmd

import (
	"errors"
	"os"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/configStructs"
	"github.com/up9inc/mizu/cli/uiUtils"
)

const analysisMessageToConfirm = `NOTE: running mizu with --analysis flag will upload recorded traffic for further analysis and enriched presentation options.`

var tapCmd = &cobra.Command{
	Use:   "tap [POD REGEX]",
	Short: "Record ingoing traffic of a kubernetes pod",
	Long: `Record the ingoing traffic of a kubernetes pod.
Supported protocols are HTTP and gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		go mizu.ReportRun("tap", mizu.Config.Tap)
		RunMizuTap()
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			mizu.Config.Tap.PodRegexStr = args[0]
		} else if len(args) > 1 {
			return errors.New("unexpected number of arguments")
		}

		if err := mizu.Config.Tap.Validate(); err != nil {
			return errormessage.FormatError(err)
		}

		mizu.Log.Infof("Mizu will store up to %s of traffic, old traffic will be cleared once the limit is reached.", mizu.Config.Tap.HumanMaxEntriesDBSize)

		if mizu.Config.Tap.Analysis {
			mizu.Log.Infof(analysisMessageToConfirm)
			if !uiUtils.AskForConfirmation("Would you like to proceed [Y/n]: ") {
				mizu.Log.Infof("You can always run mizu without analysis, aborting")
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
	tapCmd.Flags().StringP(configStructs.NamespaceTapName, "n", defaultTapConfig.Namespace, "Namespace selector")
	tapCmd.Flags().Bool(configStructs.AnalysisTapName, defaultTapConfig.Analysis, "Uploads traffic to UP9 for further analysis (Beta)")
	tapCmd.Flags().BoolP(configStructs.AllNamespacesTapName, "A", defaultTapConfig.AllNamespaces, "Tap all namespaces")
	tapCmd.Flags().StringP(configStructs.KubeConfigPathTapName, "k", defaultTapConfig.KubeConfigPath, "Path to kube-config file")
	tapCmd.Flags().StringArrayP(configStructs.PlainTextFilterRegexesTapName, "r", defaultTapConfig.PlainTextFilterRegexes, "List of regex expressions that are used to filter matching values from text/plain http bodies")
	tapCmd.Flags().Bool(configStructs.HideHealthChecksTapName, defaultTapConfig.HideHealthChecks, "hides requests with kube-probe or prometheus user-agent headers")
	tapCmd.Flags().Bool(configStructs.DisableRedactionTapName, defaultTapConfig.DisableRedaction, "Disables redaction of potentially sensitive request/response headers and body values")
	tapCmd.Flags().String(configStructs.HumanMaxEntriesDBSizeTapName, defaultTapConfig.HumanMaxEntriesDBSize, "override the default max entries db size of 200mb")
	tapCmd.Flags().String(configStructs.DirectionTapName, defaultTapConfig.Direction, "Record traffic that goes in this direction (relative to the tapped pod): in/any")
	tapCmd.Flags().Bool(configStructs.DryRunTapName, defaultTapConfig.DryRun, "Preview of all pods matching the regex, without tapping them")
}
