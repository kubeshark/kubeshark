package cmd

import (
	"errors"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/auth"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/cli/uiUtils"
	"os"
	"time"
)

const uploadTrafficMessageToConfirm = `NOTE: running mizu with --%s flag will upload recorded traffic for further analysis and enriched presentation options.`

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

		if config.Config.Auth.Token != "" {
			expiry, err := auth.GetExpiry(config.Config.Auth.Token)
			if err != nil {
				logger.Log.Errorf("failed to get expiry from token, err: %v", err)
			}

			if time.Now().After(*expiry) {
				return errors.New("token is expired, run `mizu auth login` to re-authenticate")
			}
		}

		if config.Config.Tap.Workspace != "" {
			askConfirmation(configStructs.WorkspaceTapName)

			if config.Config.Auth.Token == "" {
				return errors.New(fmt.Sprintf("--%s flag requires authentication, run `mizu auth login` to authenticate", configStructs.WorkspaceTapName))
			}
		}

		if config.Config.Tap.Analysis {
			askConfirmation(configStructs.AnalysisTapName)

			if config.Config.Auth.Token != "" {
				config.Config.Tap.Workspace = uiUtils.AskForAnswer(fmt.Sprintf("running mizu with --%s flag while logged in requires workspace, please provide workspace name: ", configStructs.AnalysisTapName))
			}
		}

		logger.Log.Infof("Mizu will store up to %s of traffic, old traffic will be cleared once the limit is reached.", config.Config.Tap.HumanMaxEntriesDBSize)

		return nil
	},
}

func askConfirmation(flagName string) {
	logger.Log.Infof(fmt.Sprintf(uploadTrafficMessageToConfirm, flagName))
	if !uiUtils.AskForConfirmation("Would you like to proceed [Y/n]: ") {
		logger.Log.Infof("You can always run mizu without %s, aborting", flagName)
		os.Exit(0)
	}
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
	tapCmd.Flags().StringP(configStructs.WorkspaceTapName, "w", defaultTapConfig.Workspace, "Uploads traffic to UP9 workspace for further analysis (requires auth)")
	tapCmd.Flags().String(configStructs.EnforcePolicyFile, defaultTapConfig.EnforcePolicyFile, "Yaml file path with policy rules")

	tapCmd.Flags().String(configStructs.EnforcePolicyFileDeprecated, defaultTapConfig.EnforcePolicyFileDeprecated, "Yaml file with policy rules")
	tapCmd.Flags().MarkDeprecated(configStructs.EnforcePolicyFileDeprecated, fmt.Sprintf("Use --%s instead", configStructs.EnforcePolicyFile))
}
