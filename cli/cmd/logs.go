package cmd

import (
	"context"
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/kubernetes"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/telemetry"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Create a zip file with logs for Github issue or troubleshoot",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("logs", config.Config.Logs)

		kubernetesProvider, err := kubernetes.NewProvider(config.Config.KubeConfigPath())
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
		ctx, _ := context.WithCancel(context.Background())

		if validationErr := config.Config.Logs.Validate(); validationErr != nil {
			return errormessage.FormatError(validationErr)
		}

		logger.Log.Debugf("Using file path %s", config.Config.Logs.FilePath())

		if dumpLogsErr := fsUtils.DumpLogs(kubernetesProvider, ctx, config.Config.Logs.FilePath()); dumpLogsErr != nil {
			logger.Log.Errorf("Failed dump logs %v", dumpLogsErr)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	defaultLogsConfig := configStructs.LogsConfig{}
	defaults.Set(&defaultLogsConfig)

	logsCmd.Flags().StringP(configStructs.FileLogsName, "f", defaultLogsConfig.FileStr, "Path for zip file (default current <pwd>\\mizu_logs.zip)")
}
