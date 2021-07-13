package database

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/debounce"
	"log"
	"mizuserver/pkg/models"
	"os"
	"strconv"
	"time"
)

const percentageOfMaxSizeBytesToPrune = 15
const defaultMaxDatabaseSizeBytes = 200 * 1000 * 1000

func StartEnforcingDatabaseSize() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Error creating filesystem watcher for db size enforcement: %v\n", err) // TODO: make fatal
		return
	}
	defer watcher.Close()

	maxEntriesDBByteSize, err := getMaxEntriesDBByteSize()
	if err != nil {
		log.Printf("Error parsing max db size: %v\n", err) // TODO: make fatal
		return
	}

	checkFileSizeDebouncer := debounce.NewDebouncer(5*time.Second, func() {
		checkFileSize(maxEntriesDBByteSize)
	})

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return // closed channel
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					checkFileSizeDebouncer.SetOn()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return // closed channel
				}
				fmt.Printf("filesystem watcher encountered error:%v\n", err)
			}
		}
	}()

	err = watcher.Add(DBPath)
	if err != nil {
		log.Printf("Error adding %s to filesystem watcher for db size enforcement: %v\n", DBPath, err) //TODO: make fatal
	}
	<-done
}

func getMaxEntriesDBByteSize() (int64, error) {
	maxEntriesDBByteSize := int64(defaultMaxDatabaseSizeBytes)
	var err error

	maxEntriesDBSizeByteSEnvVarValue := os.Getenv(shared.MaxEntriesDBSizeByteSEnvVar)
	if maxEntriesDBSizeByteSEnvVarValue != "" {
		maxEntriesDBByteSize, err = strconv.ParseInt(maxEntriesDBSizeByteSEnvVarValue, 10, 64)
	}
	return maxEntriesDBByteSize, err
}

func checkFileSize(maxSizeBytes int64) {
	fileStat, err := os.Stat(DBPath)
	if err != nil {
		fmt.Printf("Error checking %s file size: %v\n", DBPath, err)
	} else {
		if fileStat.Size() > maxSizeBytes {
			pruneOldEntries(fileStat.Size())
		}
	}
}

func pruneOldEntries(currentFileSize int64) {
	amountOfBytesToTrim := currentFileSize / (100 / percentageOfMaxSizeBytesToPrune)

	rows, err := GetEntriesTable().Limit(10000).Order("id").Rows()
	if err != nil {
		fmt.Printf("Error getting 10000 first db rows: %v\n", err)
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
			fmt.Printf("Error scanning db row: %v\n", err)
			continue
		}

		entryIdsToRemove = append(entryIdsToRemove, entry.ID)
		bytesToBeRemoved += int64(entry.EstimatedSizeBytes)
	}

	if len(entryIdsToRemove) > 0 {
		GetEntriesTable().Where(entryIdsToRemove).Delete(models.MizuEntry{})
		// VACUUM causes sqlite to shrink the db file after rows have been deleted, the db file will not shrink without this
		DB.Exec("VACUUM")
		fmt.Printf("Removed %d rows and cleared %s bytes", len(entryIdsToRemove), shared.BytesToHumanReadable(bytesToBeRemoved))
	} else {
		fmt.Printf("Found no rows to remove when pruning")
	}
}
