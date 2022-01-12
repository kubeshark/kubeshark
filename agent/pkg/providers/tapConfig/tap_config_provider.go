package tapConfig

import (
	"encoding/json"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	"io/ioutil"
	"mizuserver/pkg/models"
	"os"
	"sync"
)

const FilePath = shared.DataDirPath + "tap-config.json"

var lock = &sync.Mutex{}

var config *models.TapConfig

func Get() *models.TapConfig {
	if config == nil {
		lock.Lock()
		defer lock.Unlock()

		if config == nil {
			if content, err := ioutil.ReadFile(FilePath); err != nil {
				config = &models.TapConfig{TappedNamespaces: make(map[string]bool)}
				if !os.IsNotExist(err) {
					logger.Log.Errorf("Error loading tap config from file, err: %v", err)
				}
			} else {
				if err = json.Unmarshal(content, &config); err != nil {
					config = &models.TapConfig{TappedNamespaces: make(map[string]bool)}
					logger.Log.Errorf("Error while unmarshal tap config, err: %v", err)
				}
			}
		}
	}

	return config
}

func Save(tapConfigToSave *models.TapConfig) {
	lock.Lock()
	defer lock.Unlock()

	config = tapConfigToSave
	if data, err := json.Marshal(config); err != nil {
		logger.Log.Errorf("Error while marshal tap config, err: %v", err)
	} else {
		if err := ioutil.WriteFile(FilePath, data, 0644); err != nil {
			logger.Log.Errorf("Error writing tap config to file, err: %v", err)
		}
	}
}
