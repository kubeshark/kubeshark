package cmd

import (
	"context"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/errormessage"
	"github.com/kubeshark/kubeshark/kubeshark/fsUtils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Create a zip file with logs for Github issue or troubleshoot",
	RunE: func(cmd *cobra.Command, args []string) error {
		kubernetesProvider, err := getKubernetesProviderForCli()
		if err != nil {
			return nil
		}
		ctx := context.Background()

		if validationErr := config.Config.Logs.Validate(); validationErr != nil {
			return errormessage.FormatError(validationErr)
		}

		log.Debug().Str("logs-path", config.Config.Logs.FilePath()).Msg("Using this logs path...")

		if dumpLogsErr := fsUtils.DumpLogs(ctx, kubernetesProvider, config.Config.Logs.FilePath()); dumpLogsErr != nil {
			log.Error().Err(dumpLogsErr).Msg("Failed to dump logs.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	defaultLogsConfig := configStructs.LogsConfig{}
	if err := defaults.Set(&defaultLogsConfig); err != nil {
		log.Debug().Err(err)
	}

	logsCmd.Flags().StringP(configStructs.FileLogsName, "f", defaultLogsConfig.FileStr, "Path for zip file (default current <pwd>\\kubeshark_logs.zip)")
}
