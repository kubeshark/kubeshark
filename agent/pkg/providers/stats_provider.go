package providers

import (
	"reflect"
	"sync"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

type GeneralStats struct {
	generalStatsMutex   *sync.Mutex
	EntriesCount        int
	EntriesVolumeInGB   float64
	FirstEntryTimestamp int
	LastEntryTimestamp  int
	Buckets             map[time.Time]TimeFrameStatsValue
}

type TimeFrameStatsValue map[string]ProtocolStats

type ProtocolStats map[string]*SizeAndEntriesCount

type SizeAndEntriesCount struct {
	EntriesCount int
	BytesSize    int
}

type AccumulativeStatsCounter struct {
	Name         string `json:"name"`
	EntriesCount int    `json:"entriesCount"`
	ByteCount    int    `json:"byteCount"`
}

type AccumulativeStatsProtocol struct {
	AccumulativeStatsCounter
	Methods []*AccumulativeStatsCounter `json:"methods"`
}

var generalStats = InitGeneralStats()

func ResetGeneralStats() {
	generalStats = InitGeneralStats()
}

func InitGeneralStats() GeneralStats {
	generalStatsObj := GeneralStats{
		generalStatsMutex: &sync.Mutex{},
		Buckets:           map[time.Time]TimeFrameStatsValue{},
	}
	return generalStatsObj
}

func GetGeneralStats() GeneralStats {
	return generalStats
}

func GetAccumulativeStats() []*AccumulativeStatsProtocol {
	result := make([]*AccumulativeStatsProtocol, 0)
	for _, counters := range generalStats.Buckets {
		for protocolName, value := range counters {
			totalProtocolRequestCount := 0
			totalBytesProtocol := 0
			methods := make([]*AccumulativeStatsCounter, 0)

			for method, countersValue := range value {
				methodData := &AccumulativeStatsCounter{
					Name:         method,
					EntriesCount: 0,
					ByteCount:    0,
				}
				totalProtocolRequestCount += countersValue.EntriesCount
				methodData.EntriesCount += countersValue.EntriesCount
				totalBytesProtocol += countersValue.BytesSize
				methodData.ByteCount += countersValue.BytesSize
				methods = append(methods, methodData)
			}
			newProtocolData := &AccumulativeStatsProtocol{
				AccumulativeStatsCounter: AccumulativeStatsCounter{
					Name:         protocolName,
					EntriesCount: totalProtocolRequestCount,
					ByteCount:    totalBytesProtocol,
				},
				Methods: methods,
			}
			result = append(result, newProtocolData)
		}
	}
	return result
}

func EntryAdded(size int, summery *api.BaseEntry) {
	generalStats.generalStatsMutex.Lock()
	generalStats.EntriesCount++
	generalStats.EntriesVolumeInGB += float64(size) / (1 << 30)

	currentTimestamp := int(time.Now().Unix())

	if reflect.Value.IsZero(reflect.ValueOf(generalStats.FirstEntryTimestamp)) {
		generalStats.FirstEntryTimestamp = currentTimestamp
	}

	entryTimeBucket := time.Unix(summery.Timestamp, 0).Round(time.Minute * 5)
	if _, found := generalStats.Buckets[entryTimeBucket]; !found {
		generalStats.Buckets[entryTimeBucket] = TimeFrameStatsValue{}
	}
	if _, found := generalStats.Buckets[entryTimeBucket][summery.Protocol.Name]; !found {
		generalStats.Buckets[entryTimeBucket][summery.Protocol.Name] = ProtocolStats{}
	}
	if _, found := generalStats.Buckets[entryTimeBucket][summery.Protocol.Name][summery.Method]; !found {
		generalStats.Buckets[entryTimeBucket][summery.Protocol.Name][summery.Method] = &SizeAndEntriesCount{
			BytesSize:    0,
			EntriesCount: 0,
		}
	}

	generalStats.Buckets[entryTimeBucket][summery.Protocol.Name][summery.Method].EntriesCount += 1
	generalStats.Buckets[entryTimeBucket][summery.Protocol.Name][summery.Method].BytesSize += size

	generalStats.LastEntryTimestamp = currentTimestamp
	generalStats.generalStatsMutex.Lock()
}
