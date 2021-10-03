package cmd

import (
	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
	"github.com/up9inc/mizu/cli/auth"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/telemetry"
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to up9 application",
	Annotations: map[string]string{
		"ConfigSection": "auth",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		go telemetry.ReportRun("authLogin", config.Config.Auth)

		if config.Config.Auth.Token != "" {
			logger.Log.Infof("already logged in")
			return nil
		}

		if err := config.Config.Auth.Validate(); err != nil {
			return errormessage.FormatError(err)
		}

		if config.Config.Auth.ClientId == "" && config.Config.Auth.ClientSecret == "" {
			logger.Log.Infof("temporary")
			return nil
		}

		token, err := auth.LoginNonInteractively(config.Config.Auth.ClientId, config.Config.Auth.ClientSecret, config.Config.Auth.EnvName)
		if err != nil {
			logger.Log.Errorf("Failed creating token, err: %v", err)
			return nil
		}

		configWithDefaults, err := config.GetConfigWithDefaults()
		if err != nil {
			logger.Log.Errorf("Failed generating config with defaults, err: %v", err)
			return nil
		}

		authConfig := configStructs.AuthConfig{
			EnvName:      config.Config.Auth.EnvName,
			ClientId:     config.Config.Auth.ClientId,
			ClientSecret: config.Config.Auth.ClientSecret,
			Token:        token.AccessToken,
			ExpiryDate:   token.Expiry,
		}

		configWithDefaults.Auth = authConfig

		if err := config.WriteConfig(configWithDefaults); err != nil {
			logger.Log.Errorf("Failed writing config with auth, err: %v", err)
			return nil
		}

		logger.Log.Infof("Login completed")

		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)

	defaultAuthConfig := configStructs.AuthConfig{}
	defaults.Set(&defaultAuthConfig)

	authLoginCmd.Flags().String(configStructs.ClientIdAuthName,  defaultAuthConfig.ClientId, "Client Id")
	authLoginCmd.Flags().String(configStructs.ClientSecretAuthName,  defaultAuthConfig.ClientSecret, "Client Secret")
}
