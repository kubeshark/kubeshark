package tappers

import (
	"os"
	"sync"

	"github.com/up9inc/mizu/agent/pkg/utils"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
)

const FilePath = shared.DataDirPath + "tappers-status.json"

var (
	lockStatus = &sync.Mutex{}
	syncOnce   sync.Once
	status     map[string]*shared.TapperStatus

	lockConnectedCount = &sync.Mutex{}
	connectedCount     int
)

func GetStatus() map[string]*shared.TapperStatus {
	initStatus()

	return status
}

func SetStatus(tapperStatus *shared.TapperStatus) {
	initStatus()

	lockStatus.Lock()
	defer lockStatus.Unlock()

	status[tapperStatus.NodeName] = tapperStatus

	saveStatus()
}

func ResetStatus() {
	lockStatus.Lock()
	defer lockStatus.Unlock()

	status = make(map[string]*shared.TapperStatus)

	saveStatus()
}

func GetConnectedCount() int {
	return connectedCount
}

func Connected() {
	lockConnectedCount.Lock()
	defer lockConnectedCount.Unlock()

	connectedCount++
}

func Disconnected() {
	lockConnectedCount.Lock()
	defer lockConnectedCount.Unlock()

	connectedCount--
}

func initStatus() {
	syncOnce.Do(func() {
		if err := utils.ReadJsonFile(FilePath, &status); err != nil {
			status = make(map[string]*shared.TapperStatus)

			if !os.IsNotExist(err) {
				logger.Log.Errorf("Error reading tappers status from file, err: %v", err)
			}
		}
	})
}

func saveStatus() {
	if err := utils.SaveJsonFile(FilePath, status); err != nil {
		logger.Log.Errorf("Error saving tappers status, err: %v", err)
	}
}
