package providers_test

import (
	"fmt"
	"mizuserver/pkg/database"
	"mizuserver/pkg/providers"
	"testing"
)

const dbPath = "./entries.db"

func init() {
	database.InitDataBase(dbPath)
}

func TestNoEntryAddedCount(t *testing.T) {
	entriesStats := providers.GetGeneralStats()

	if entriesStats.EntriesCount != 0 {
		t.Errorf("unexpected result - expected: %v, actual: %v", 0, entriesStats.EntriesCount)
	}
}

func TestEntryAddedCount(t *testing.T) {
	tests := []int{1, 5, 10, 100, 500, 1000}

	for _, entriesCount := range tests {
		t.Run(fmt.Sprintf("%d", entriesCount), func(t *testing.T) {
			for i := 0; i < entriesCount; i++ {
				providers.EntryAdded()
			}

			entriesStats := providers.GetGeneralStats()

			if entriesStats.EntriesCount != entriesCount {
				t.Errorf("unexpected result - expected: %v, actual: %v", entriesCount, entriesStats.EntriesCount)
			}

			t.Cleanup(providers.ResetGeneralStats)
		})
	}
}
