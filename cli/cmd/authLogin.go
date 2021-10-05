package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/auth"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/telemetry"
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to up9 application",
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("authLogin", config.Config.Auth)

		token, err := auth.LoginInteractively(config.Config.Auth.EnvName)
		if err != nil {
			logger.Log.Errorf("Failed login interactively, err: %v", err)
			return nil
		}

		authConfig := configStructs.AuthConfig{
			EnvName:      config.Config.Auth.EnvName,
			Token:        token.AccessToken,
			ExpiryDate:   token.Expiry,
		}

		config.Config.Auth = authConfig

		if err := config.WriteConfig(&config.Config); err != nil {
			logger.Log.Errorf("Failed writing config with auth, err: %v", err)
			return nil
		}

		logger.Log.Infof("Login successfully, token stored in config")

		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
}
