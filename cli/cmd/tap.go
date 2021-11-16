package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/auth"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/telemetry"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
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

		if config.Config.Tap.Workspace != "" {
			askConfirmation(configStructs.WorkspaceTapName)

			if config.Config.Auth.Token == "" {
				logger.Log.Infof("This action requires authentication, please log in to continue")
				if err := auth.Login(); err != nil {
					logger.Log.Errorf("failed to log in, err: %v", err)
					return nil
				}
			} else {
				tokenExpired, err := shared.IsTokenExpired(config.Config.Auth.Token)
				if err != nil {
					logger.Log.Errorf("failed to check if token is expired, err: %v", err)
					return nil
				}

				if tokenExpired {
					logger.Log.Infof("Token expired, please log in again to continue")
					if err := auth.Login(); err != nil {
						logger.Log.Errorf("failed to log in, err: %v", err)
						return nil
					}
				}
			}
		}

		if config.Config.Tap.Analysis {
			askConfirmation(configStructs.AnalysisTapName)

			config.Config.Auth.Token = ""
		}

		logger.Log.Infof("Mizu will store up to %s of traffic, old traffic will be cleared once the limit is reached.", config.Config.Tap.HumanMaxEntriesDBSize)

		return nil
	},
}

func askConfirmation(flagName string) {
	logger.Log.Infof(fmt.Sprintf(uploadTrafficMessageToConfirm, flagName))

	if !config.Config.Tap.AskUploadConfirmation {
		return
	}

	if !uiUtils.AskForConfirmation("Would you like to proceed [Y/n]: ") {
		logger.Log.Infof("You can always run mizu without %s, aborting", flagName)
		os.Exit(0)
	}

	if err := config.UpdateConfig(func(configStruct *config.ConfigStruct) { configStruct.Tap.AskUploadConfirmation = false }); err != nil {
		logger.Log.Debugf("failed updating config with upload confirmation, err: %v", err)
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
	tapCmd.Flags().StringP(configStructs.WorkspaceTapName, "w", defaultTapConfig.Workspace, "Uploads traffic to your UP9 workspace for further analysis (requires auth)")
	tapCmd.Flags().String(configStructs.EnforcePolicyFile, defaultTapConfig.EnforcePolicyFile, "Yaml file path with policy rules")
	tapCmd.Flags().String(configStructs.ContractFile, defaultTapConfig.ContractFile, "OAS/Swagger file to validate to monitor the contracts")
	tapCmd.Flags().Bool(configStructs.DaemonModeTapName, defaultTapConfig.DaemonMode, "Run mizu in daemon mode, detached from the cli")
	tapCmd.Flags().Bool(configStructs.IstioName, defaultTapConfig.Istio, "Record decrypted traffic if the cluster configured with istio and mtls")
}
