package providers_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/up9inc/mizu/agent/pkg/providers"
	"github.com/up9inc/mizu/tap/api"
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

	entryBucketKey := time.Date(2021, 1, 1, 10, 0, 0, 0, time.UTC)
	valueLessThanBucketThreshold := time.Second * 130
	mockSummery := &api.BaseEntry{Protocol: api.Protocol{Name: "mock"}, Method: "mock-method", Timestamp: entryBucketKey.Add(valueLessThanBucketThreshold).UnixNano()}
	for _, entriesCount := range tests {
		t.Run(fmt.Sprintf("%d", entriesCount), func(t *testing.T) {
			for i := 0; i < entriesCount; i++ {
				providers.EntryAdded(0, mockSummery)
			}

			entriesStats := providers.GetGeneralStats()

			if entriesStats.EntriesCount != entriesCount {
				t.Errorf("unexpected result - expected: %v, actual: %v", entriesCount, entriesStats.EntriesCount)
			}

			if entriesStats.EntriesVolumeInGB != 0 {
				t.Errorf("unexpected result - expected: %v, actual: %v", 0, entriesStats.EntriesVolumeInGB)
			}

			t.Cleanup(func() {
				providers.ResetGeneralStats()
				generalStats := providers.GetGeneralStats()
				if generalStats.EntriesCount != 0 {
					t.Errorf("unexpected result - expected: %v, actual: %v", 0, generalStats.EntriesCount)
				}

			})
		})
	}
}

func TestEntryAddedVolume(t *testing.T) {
	// 6 bytes + 4 bytes
	tests := [][]byte{[]byte("volume"), []byte("test")}
	var expectedEntriesCount int
	var expectedVolumeInGB float64

	mockSummery := &api.BaseEntry{Protocol: api.Protocol{Name: "mock"}, Method: "mock-method", Timestamp: time.Date(2021, 1, 1, 10, 0, 0, 0, time.UTC).UnixNano()}

	for _, data := range tests {
		t.Run(fmt.Sprintf("%d", len(data)), func(t *testing.T) {
			expectedEntriesCount++
			expectedVolumeInGB += float64(len(data)) / (1 << 30)

			providers.EntryAdded(len(data), mockSummery)

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
