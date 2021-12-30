package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mizuserver/pkg/config"
	"mizuserver/pkg/models"
	"os"
	"sync"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

const (
	installStateFileName = "install_state.json"
	adminUser            = "admin"
)

var (
	installStateFilePath = fmt.Sprintf("%s%s", shared.ConfigDirPath, installStateFileName)
	lock                 = sync.RWMutex{}
)

func IsInstallNeeded() (bool, error) {
	if config.Config.IsEphermeral {
		return false, nil
	}

	lock.Lock()
	defer lock.Unlock()

	if _, err := os.Stat(installStateFilePath); os.IsNotExist(err) {
		return true, nil
	} else {
		install_state := &models.InstallState{}
		if install_state_json, err := ioutil.ReadFile(installStateFilePath); err != nil {
			return true, err
		} else if err = json.Unmarshal(install_state_json, install_state); err != nil {
			return true, err
		}

		return !install_state.Completed, nil
	}
}

func DoInstall(adminPassword string, ctx context.Context) error {
	lock.Lock()
	defer lock.Unlock()

	if IsInstallNeeded, err := IsInstallNeeded(); err != nil {
		return err
	} else if !IsInstallNeeded {
		return errors.New("install has already been performed")
	}

	_, adminIdentityId, err := RegisterUser(adminUser, adminPassword, ctx)
	if err != nil {
		return err
	}

	err = createInstallStateFile(true)
	if err != nil {
		revertInstall(adminIdentityId, ctx)
		return err
	}

	return nil
}

func revertInstall(adminIdentitiyId string, ctx context.Context) {
	err := os.Remove(installStateFilePath)
	if err != nil {
		logger.Log.Errorf("error occured while removing install state file: %v", err)
	}

	if adminIdentitiyId != "" {
		err = DeleteUser(adminIdentitiyId, ctx)
		if err != nil {
			logger.Log.Errorf("error occured while removing admin user: %v", err)
		}
	}
}

func createInstallStateFile(state bool) error {
	install_state := &models.InstallState{
		Completed: state,
	}

	if install_state_json, err := json.Marshal(install_state); err != nil {
		return err
	} else {
		return ioutil.WriteFile(installStateFilePath, install_state_json, 0644)
	}
}
