package cmd

import (
	"context"
	"fmt"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/misc/fsUtils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Create a ZIP file with logs for GitHub issues or troubleshooting",
	RunE: func(cmd *cobra.Command, args []string) error {
		kubernetesProvider, err := getKubernetesProviderForCli(false, false)
		if err != nil {
			return nil
		}
		ctx := context.Background()

		if validationErr := config.Config.Logs.Validate(); validationErr != nil {
			return errormessage.FormatError(validationErr)
		}

		log.Debug().Str("logs-path", config.Config.Logs.FilePath()).Msg("Using this logs path...")

		if dumpLogsErr := fsUtils.DumpLogs(ctx, kubernetesProvider, config.Config.Logs.FilePath(), config.Config.Logs.Grep); dumpLogsErr != nil {
			log.Error().Err(dumpLogsErr).Msg("Failed to dump logs.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	defaultLogsConfig := configStructs.LogsConfig{}
	if err := defaults.Set(&defaultLogsConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	logsCmd.Flags().StringP(configStructs.FileLogsName, "f", defaultLogsConfig.FileStr, fmt.Sprintf("Path for zip file (default current <pwd>\\%s_logs.zip)", misc.Program))
	logsCmd.Flags().StringP(configStructs.GrepLogsName, "g", defaultLogsConfig.Grep, "Regexp to do grepping on the logs")
}
