package providers

import (
	"reflect"
	"time"
)

type GeneralStats struct {
	EntriesCount        int
	FirstEntryTimestamp int
	LastEntryTimestamp  int
}

var (
	generalStats GeneralStats
)

func init() {
	generalStats = GeneralStats{}
}

func ResetGeneralStats() {
	generalStats = GeneralStats{}
}

func GetGeneralStats() GeneralStats {
	return generalStats
}

func EntryAdded() {
	generalStats.EntriesCount++

	currentTimestamp := int(time.Now().Unix())

	if reflect.Value.IsZero(reflect.ValueOf(generalStats.FirstEntryTimestamp)) {
		generalStats.FirstEntryTimestamp = currentTimestamp
	}

	generalStats.LastEntryTimestamp = currentTimestamp
}


