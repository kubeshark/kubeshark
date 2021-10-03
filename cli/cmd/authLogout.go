package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/telemetry"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from up9 application",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("authLogout", config.Config.Auth)

		configWithDefaults, err := config.GetConfigWithDefaults()
		if err != nil {
			logger.Log.Errorf("Failed generating config with defaults, err: %v", err)
			return nil
		}

		if err := config.WriteConfig(configWithDefaults); err != nil {
			logger.Log.Errorf("Failed writing config with default auth, err: %v", err)
			return nil
		}

		logger.Log.Infof("Logout completed")

		return nil
	},
}

func init() {
	authCmd.AddCommand(authLogoutCmd)
}
