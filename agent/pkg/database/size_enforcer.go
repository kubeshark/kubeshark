package database

import (
	"github.com/fsnotify/fsnotify"
	"github.com/romana/rlog"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	"github.com/up9inc/mizu/shared/units"
	"log"
	"mizuserver/pkg/models"
	"os"
	"strconv"
	"time"
)

const percentageOfMaxSizeBytesToPrune = 15
const defaultMaxDatabaseSizeBytes int64 = 200 * 1000 * 1000

func StartEnforcingDatabaseSize() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating filesystem watcher for db size enforcement: %v\n", err)
		return
	}

	maxEntriesDBByteSize, err := getMaxEntriesDBByteSize()
	if err != nil {
		log.Fatalf("Error parsing max db size: %v\n", err)
		return
	}

	checkFileSizeDebouncer := debounce.NewDebouncer(5*time.Second, func() {
		checkFileSize(maxEntriesDBByteSize)
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
				rlog.Errorf("filesystem watcher encountered error:%v", err)
			}
		}
	}()

	err = watcher.Add(DBPath)
	if err != nil {
		log.Fatalf("Error adding %s to filesystem watcher for db size enforcement: %v\n", DBPath, err)
	}
}

func getMaxEntriesDBByteSize() (int64, error) {
	maxEntriesDBByteSize := defaultMaxDatabaseSizeBytes
	var err error

	maxEntriesDBSizeByteSEnvVarValue := os.Getenv(shared.MaxEntriesDBSizeBytesEnvVar)
	if maxEntriesDBSizeByteSEnvVarValue != "" {
		maxEntriesDBByteSize, err = strconv.ParseInt(maxEntriesDBSizeByteSEnvVarValue, 10, 64)
	}
	return maxEntriesDBByteSize, err
}

func checkFileSize(maxSizeBytes int64) {
	fileStat, err := os.Stat(DBPath)
	if err != nil {
		rlog.Errorf("Error checking %s file size: %v", DBPath, err)
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
		rlog.Errorf("Error getting 10000 first db rows: %v", err)
		return
	}

	entryIdsToRemove := make([]uint, 0)
	bytesToBeRemoved := int64(0)
	for rows.Next() {
		if bytesToBeRemoved >= amountOfBytesToTrim {
			break
		}
		var entry models.MizuEntry
		err = DB.ScanRows(rows, &entry)
		if err != nil {
			rlog.Errorf("Error scanning db row: %v", err)
			continue
		}

		entryIdsToRemove = append(entryIdsToRemove, entry.ID)
		bytesToBeRemoved += int64(entry.EstimatedSizeBytes)
	}

	if len(entryIdsToRemove) > 0 {
		GetEntriesTable().Where(entryIdsToRemove).Delete(models.MizuEntry{})
		// VACUUM causes sqlite to shrink the db file after rows have been deleted, the db file will not shrink without this
		DB.Exec("VACUUM")
		rlog.Errorf("Removed %d rows and cleared %s", len(entryIdsToRemove), units.BytesToHumanReadable(bytesToBeRemoved))
	} else {
		rlog.Error("Found no rows to remove when pruning")
	}
}
