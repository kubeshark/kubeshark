package database

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/up9inc/mizu/shared/debounce"
	"log"
	"mizuserver/pkg/models"
	"os"
	"time"
)

const percentageOfMaxSizeBytesToPrune = 5

func StartEnforcingDatabaseSize(maxSizeBytes int64) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating filesystem watcher for db size enforcement:", err)
	}
	defer watcher.Close()

	checkFileSizeDebouncer := debounce.NewDebouncer(5 * time.Second, func() {
		checkFileSize(maxSizeBytes)
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
		log.Fatal(fmt.Sprintf("Error adding %s to filesystem watcher for db size enforcement: %v", DBPath, err))
	}
	<-done
}

func checkFileSize(maxSizeBytes int64) {
	fileStat, err := os.Stat(DBPath)
	if err != nil {
		fmt.Printf("Error checking %s file size: %v\n", DBPath, err)
	} else {
		if fileStat.Size() > maxSizeBytes {
			pruneOldEntries(maxSizeBytes)
		}
	}
}

func pruneOldEntries(maxSizeBytes int64) {
	amountOfBytesToTrim := maxSizeBytes * (percentageOfMaxSizeBytesToPrune / 100)

	rows, err := GetEntriesTable().Limit(100).Order("id").Rows()
	if err != nil {
		fmt.Printf("Error getting 100 first db rows: %v\n", err)
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
		GetEntriesTable().Delete(entryIdsToRemove)
		fmt.Printf("Removed %d rows and cleared %d bytes", len(entryIdsToRemove), bytesToBeRemoved)
	}
}
