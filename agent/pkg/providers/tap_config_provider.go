package providers

import (
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	"io/ioutil"
	"mizuserver/pkg/models"
	"os"
	"sync"
)

const TapConfigFileName = "tap-config.json"

var configLock = &sync.Mutex{}

var tapConfig *models.TapConfig

func GetTapConfig() *models.TapConfig {
	if tapConfig == nil {
		configLock.Lock()
		defer configLock.Unlock()

		if tapConfig == nil {
			filePath := fmt.Sprintf("%s%s", shared.DataDirPath, TapConfigFileName)

			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				tapConfig = &models.TapConfig{TappedNamespaces: make(map[string]bool)}
				if !os.IsNotExist(err) {
					logger.Log.Errorf("Error loading tap config from file, err: %v", err)
				}
			} else {
				if err = json.Unmarshal(content, &tapConfig); err != nil {
					tapConfig = &models.TapConfig{TappedNamespaces: make(map[string]bool)}
					logger.Log.Errorf("Error while unmarshal tap config, err: %v", err)
				}
			}
		}
	}

	return tapConfig
}

func SaveTapConfig(tapConfigToSave *models.TapConfig) {
	configLock.Lock()
	defer configLock.Unlock()

	tapConfig = tapConfigToSave
	data, err := json.Marshal(tapConfig)
	if err != nil {
		logger.Log.Errorf("Error while marshal tap config, err: %v", err)
	} else {
		filePath := fmt.Sprintf("%s%s", shared.DataDirPath, TapConfigFileName)
		if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
			logger.Log.Errorf("Error writing tap config to file, err: %v", err)
		}
	}
}
