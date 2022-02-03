package providers_test

import (
	"fmt"
	"testing"

	"github.com/up9inc/mizu/agent/pkg/providers"
)

func TestNoEntryAddedCount(t *testing.T) {
	entriesStats := providers.GetGeneralStats()

	if entriesStats.EntriesCount != 0 {
		t.Errorf("unexpected result - expected: %v, actual: %v", 0, entriesStats.EntriesCount)
	}

	if entriesStats.EntriesVolumeInGB != 0 {
		t.Errorf("unexpected result - expected: %v, actual: %v", 0, entriesStats.EntriesVolumeInGB)
	}
}

func TestEntryAddedCount(t *testing.T) {
	tests := []int{1, 5, 10, 100, 500, 1000}

	for _, entriesCount := range tests {
		t.Run(fmt.Sprintf("%d", entriesCount), func(t *testing.T) {
			for i := 0; i < entriesCount; i++ {
				providers.EntryAdded(0)
			}

			entriesStats := providers.GetGeneralStats()

			if entriesStats.EntriesCount != entriesCount {
				t.Errorf("unexpected result - expected: %v, actual: %v", entriesCount, entriesStats.EntriesCount)
			}

			if entriesStats.EntriesVolumeInGB != 0 {
				t.Errorf("unexpected result - expected: %v, actual: %v", 0, entriesStats.EntriesVolumeInGB)
			}

			t.Cleanup(providers.ResetGeneralStats)
		})
	}
}

func TestEntryAddedVolume(t *testing.T) {
	// 6 bytes + 4 bytes
	tests := [][]byte{[]byte("volume"), []byte("test")}
	var expectedEntriesCount int
	var expectedVolumeInGB float64

	for _, data := range tests {
		t.Run(fmt.Sprintf("%d", len(data)), func(t *testing.T) {
			expectedEntriesCount++
			expectedVolumeInGB += float64(len(data)) / (1 << 30)

			providers.EntryAdded(len(data))

			entriesStats := providers.GetGeneralStats()

			if entriesStats.EntriesCount != expectedEntriesCount {
				t.Errorf("unexpected result - expected: %v, actual: %v", expectedEntriesCount, entriesStats.EntriesCount)
			}

			if entriesStats.EntriesVolumeInGB != expectedVolumeInGB {
				t.Errorf("unexpected result - expected: %v, actual: %v", expectedVolumeInGB, entriesStats.EntriesVolumeInGB)
			}
		})
	}

}
