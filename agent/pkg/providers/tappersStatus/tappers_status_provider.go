package tappersStatus

import (
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	"mizuserver/pkg/utils"
	"os"
	"sync"
)

const FilePath = shared.DataDirPath + "tappers-status.json"

var (
	lock           = &sync.Mutex{}
	syncOnce       sync.Once
	syncOnceCalled = false
	tappersStatus  map[string]*shared.TapperStatus
)

func Get() map[string]*shared.TapperStatus {
	syncOnce.Do(func() {
		if err := utils.ReadJsonFile(FilePath, &tappersStatus); err != nil {
			tappersStatus = make(map[string]*shared.TapperStatus)

			if !os.IsNotExist(err) {
				logger.Log.Errorf("Error reading tappers status from file, err: %v", err)
			}
		}

		syncOnceCalled = true
	})

	return tappersStatus
}

func Set(tapperStatus *shared.TapperStatus) {
	if !syncOnceCalled {
		Get() // make sure we get tappers status value from file if it exists
	}

	lock.Lock()
	defer lock.Unlock()

	tappersStatus[tapperStatus.NodeName] = tapperStatus

	save()
}

func Reset() {
	lock.Lock()
	defer lock.Unlock()

	tappersStatus = make(map[string]*shared.TapperStatus)

	save()
}

func save() {
	if err := utils.SaveJsonFile(FilePath, tappersStatus); err != nil {
		logger.Log.Errorf("Error saving tappers status, err: %v", err)
	}
}
