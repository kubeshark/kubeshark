package providers

import (
	"context"
	"mizuserver/pkg/config"
)

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
