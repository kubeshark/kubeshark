package cmd

import (
	"errors"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/units"
	"os"
	"regexp"
	"strings"
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
		cmd.Flags().Visit(mizu.InitFlag)

		if len(args) == 1 {
			mizu.Config.Tap.PodRegexStr = args[0]
		} else if len(args) > 1 {
			return errors.New("unexpected number of arguments")
		}

		var compileErr error
		mizu.Config.Tap.PodRegex, compileErr = regexp.Compile(mizu.Config.Tap.PodRegexStr)
		if compileErr != nil {
			return errors.New(fmt.Sprintf("%s is not a valid regex %s", mizu.Config.Tap.PodRegexStr, compileErr))
		}

		var parseHumanDataSizeErr error
		mizu.Config.Tap.MaxEntriesDBSizeBytes, parseHumanDataSizeErr = units.HumanReadableToBytes(mizu.Config.Tap.HumanMaxEntriesDBSize)
		if parseHumanDataSizeErr != nil {
			return errors.New(fmt.Sprintf("Could not parse --max-entries-db-size value %s", mizu.Config.Tap.HumanMaxEntriesDBSize))
		}
		mizu.Log.Infof("Mizu will store up to %s of traffic, old traffic will be cleared once the limit is reached.", units.BytesToHumanReadable(mizu.Config.Tap.MaxEntriesDBSizeBytes))

		directionLowerCase := strings.ToLower(mizu.Config.Tap.Direction)
		if directionLowerCase == "any" {
			mizu.Config.Tap.TapOutgoing = true
		} else if directionLowerCase == "in" {
			mizu.Config.Tap.TapOutgoing = false
		} else {
			return errors.New(fmt.Sprintf("%s is not a valid value for flag --direction. Acceptable values are in/any.", mizu.Config.Tap.Direction))
		}

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

	defaultTapConfig := mizu.TapConfig{}
	defaults.Set(&defaultTapConfig)

	tapCmd.Flags().Uint16P("gui-port", "p", defaultTapConfig.GuiPort, "Provide a custom port for the web interface webserver")
	tapCmd.Flags().StringP("namespace", "n", defaultTapConfig.Namespace, "Namespace selector")
	tapCmd.Flags().Bool("analysis", defaultTapConfig.Analysis, "Uploads traffic to UP9 for further analysis (Beta)")
	tapCmd.Flags().BoolP("all-namespaces", "A", defaultTapConfig.AllNamespaces, "Tap all namespaces")
	tapCmd.Flags().StringP("kube-config", "k", defaultTapConfig.KubeConfigPath, "Path to kube-config file")
	tapCmd.Flags().StringArrayP("regex-masking", "r", defaultTapConfig.PlainTextFilterRegexes, "List of regex expressions that are used to filter matching values from text/plain http bodies")
	tapCmd.Flags().Bool("hide-healthchecks", defaultTapConfig.HideHealthChecks, "hides requests with kube-probe or prometheus user-agent headers")
	tapCmd.Flags().Bool("no-redact", defaultTapConfig.DisableRedaction, "Disables redaction of potentially sensitive request/response headers and body values")
	tapCmd.Flags().String("max-entries-db-size", defaultTapConfig.HumanMaxEntriesDBSize, "override the default max entries db size of 200mb")
	tapCmd.Flags().String("direction", defaultTapConfig.Direction, "Record traffic that goes in this direction (relative to the tapped pod): in/any")
}
