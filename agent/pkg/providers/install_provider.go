package providers

import (
	"context"
	"errors"
	"mizuserver/pkg/config"

	"github.com/up9inc/mizu/shared/logger"
)

const (
	installStateFileName = "install_state.json"
	adminUser            = "admin"
)

func IsInstallNeeded() (bool, error) {
	logger.Log.Infof("config.Config.RequireUserAuth %b", config.Config.RequireUserAuth)
	if !config.Config.RequireUserAuth {
		return false, nil
	}

	if anyUserExists, err := AnyUserExists(context.Background()); err != nil {
		return false, err
	} else {
		return !anyUserExists, nil
	}
}

func DoInstall(adminPassword string, ctx context.Context) (*string, error) {
	if IsInstallNeeded, err := IsInstallNeeded(); err != nil {
		return nil, err
	} else if !IsInstallNeeded {
		return nil, errors.New("install has already been performed")
	}

	token, _, err := RegisterUser(adminUser, adminPassword, ctx)
	if err != nil {
		return nil, err
	}

	return token, nil
}
