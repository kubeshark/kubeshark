package providers

import (
	"context"
	"errors"
	"mizuserver/pkg/config"

	ory "github.com/ory/kratos-client-go"
)

const AdminUsername = "admin"

func IsInstallNeeded() (bool, error) {
	if !config.Config.StandaloneMode { // install not needed in ephermeral mizu
		return false, nil
	}

	if anyUserExists, err := AnyUserExists(context.Background()); err != nil {
		return false, err
	} else {
		return !anyUserExists, nil
	}
}

func CreateAdminUser(password string, ctx context.Context) (token *string, err error, formErrorMessages map[string][]ory.UiText) {
	if isInstallNeeded, err := IsInstallNeeded(); err != nil {
		return nil, err, nil
	} else if !isInstallNeeded {
		return nil, errors.New("The admin user has already been created"), nil
	}

	token, identityId, err, formErrors := RegisterUser(AdminUsername, password, ctx)
	if err != nil {
		return nil, err, formErrors
	}

	err = SetUserSystemRole(AdminUsername, AdminRole)

	if err != nil {
		//Delete the user to prevent a half-setup situation where admin user is created without admin privileges
		DeleteUser(identityId, ctx)

		return nil, err, nil
	}

	return token, nil, nil
}
