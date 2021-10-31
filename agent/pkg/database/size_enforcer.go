package database

import (
	"mizuserver/pkg/config"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/up9inc/mizu/shared/debounce"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/shared/units"
	tapApi "github.com/up9inc/mizu/tap/api"
)

const percentageOfMaxSizeBytesToPrune = 15

func StartEnforcingDatabaseSize() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Log.Fatalf("Error creating filesystem watcher for db size enforcement: %v\n", err)
		return
	}

	checkFileSizeDebouncer := debounce.NewDebouncer(5*time.Second, func() {
		checkFileSize(config.Config.MaxDBSizeBytes)
	})

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return // closed channel
				}
				if event.Op == fsnotify.Write {
					checkFileSizeDebouncer.SetOn()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return // closed channel
				}
				logger.Log.Errorf("filesystem watcher encountered error:%v", err)
			}
		}
	}()

	err = watcher.Add(DBPath)
	if err != nil {
		logger.Log.Fatalf("Error adding %s to filesystem watcher for db size enforcement: %v\n", DBPath, err)
	}
}

func checkFileSize(maxSizeBytes int64) {
	fileStat, err := os.Stat(DBPath)
	if err != nil {
		logger.Log.Errorf("Error checking %s file size: %v", DBPath, err)
	} else {
		if fileStat.Size() > maxSizeBytes {
			pruneOldEntries(fileStat.Size())
		}
	}
}

func pruneOldEntries(currentFileSize int64) {
	// sqlite locks the database while delete or VACUUM are running and sqlite is terrible at handling its own db lock while a lot of inserts are attempted, we prevent a significant bottleneck by handling the db lock ourselves here
	IsDBLocked = true
	defer func() { IsDBLocked = false }()

	amountOfBytesToTrim := currentFileSize / (100 / percentageOfMaxSizeBytesToPrune)

	rows, err := GetEntriesTable().Limit(10000).Order("id").Rows()
	if err != nil {
		logger.Log.Errorf("Error getting 10000 first db rows: %v", err)
		return
	}

	entryIdsToRemove := make([]uint, 0)
	bytesToBeRemoved := int64(0)
	for rows.Next() {
		if bytesToBeRemoved >= amountOfBytesToTrim {
			break
		}
		var entry tapApi.MizuEntry
		err = DB.ScanRows(rows, &entry)
		if err != nil {
			logger.Log.Errorf("Error scanning db row: %v", err)
			continue
		}

		entryIdsToRemove = append(entryIdsToRemove, entry.ID)
		bytesToBeRemoved += int64(entry.EstimatedSizeBytes)
	}

	if len(entryIdsToRemove) > 0 {
		GetEntriesTable().Where(entryIdsToRemove).Delete(tapApi.MizuEntry{})
		// VACUUM causes sqlite to shrink the db file after rows have been deleted, the db file will not shrink without this
		DB.Exec("VACUUM")
		logger.Log.Errorf("Removed %d rows and cleared %s", len(entryIdsToRemove), units.BytesToHumanReadable(bytesToBeRemoved))
	} else {
		logger.Log.Error("Found no rows to remove when pruning")
	}
}
