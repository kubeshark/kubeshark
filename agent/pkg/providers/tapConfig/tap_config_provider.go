package tapConfig

import (
	"os"
	"sync"

	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/agent/pkg/utils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
)

const FilePath = shared.DataDirPath + "tap-config.json"

var (
	lock     = &sync.Mutex{}
	syncOnce sync.Once
	config   *models.TapConfig
)

func Get() *models.TapConfig {
	syncOnce.Do(func() {
		if err := utils.ReadJsonFile(FilePath, &config); err != nil {
			config = &models.TapConfig{TappedNamespaces: make(map[string]bool)}

			if !os.IsNotExist(err) {
				logger.Log.Errorf("Error reading tap config from file, err: %v", err)
			}
		}
	})

	return config
}

func Save(tapConfigToSave *models.TapConfig) {
	lock.Lock()
	defer lock.Unlock()

	config = tapConfigToSave
	if err := utils.SaveJsonFile(FilePath, config); err != nil {
		logger.Log.Errorf("Error saving tap config, err: %v", err)
	}
}
