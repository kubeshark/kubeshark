package providers

import (
	"reflect"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

type GeneralStats struct {
	EntriesCount        int
	EntriesVolumeInGB   float64
	FirstEntryTimestamp int
	LastEntryTimestamp  int
}

type TimeFrameStatsValue map[string]ProtocolStats

type ProtocolStats struct {
	MethodsStats map[string]*SizeAndEntriesCount
	Color        string
}

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
	Color   string                      `json:"color"`
	Methods []*AccumulativeStatsCounter `json:"methods"`
}

type BucketStats map[time.Time]TimeFrameStatsValue

var (
	generalStats = GeneralStats{}
	bucketsStats = BucketStats{}
)

func ResetGeneralStats() {
	generalStats = GeneralStats{}
}

func GetGeneralStats() GeneralStats {
	return generalStats
}

func GetAccumulativeStats() []*AccumulativeStatsProtocol {
	result := make([]*AccumulativeStatsProtocol, 0)
	for _, counters := range bucketsStats {
		for protocolName, value := range counters {
			totalProtocolRequestCount := 0
			totalBytesProtocol := 0
			methods := make([]*AccumulativeStatsCounter, 0)

			for method, countersValue := range value.MethodsStats {
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
	generalStats.EntriesCount++
	generalStats.EntriesVolumeInGB += float64(size) / (1 << 30)

	currentTimestamp := int(time.Now().Unix())

	if reflect.Value.IsZero(reflect.ValueOf(generalStats.FirstEntryTimestamp)) {
		generalStats.FirstEntryTimestamp = currentTimestamp
	}

	addToBucketStats(size, summery)

	generalStats.LastEntryTimestamp = currentTimestamp
}

func addToBucketStats(size int, summery *api.BaseEntry) {
	entryTimeBucketRounded := time.Unix(summery.Timestamp, 0).Round(time.Minute * 5)
	if _, found := bucketsStats[entryTimeBucketRounded]; !found {
		bucketsStats[entryTimeBucketRounded] = TimeFrameStatsValue{}
	}
	if _, found := bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation]; !found {
		bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation] = ProtocolStats{
			MethodsStats: map[string]*SizeAndEntriesCount{},
			Color:        summery.Protocol.BackgroundColor,
		}
	}
	if _, found := bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation].MethodsStats[summery.Method]; !found {
		bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation].MethodsStats[summery.Method] = &SizeAndEntriesCount{
			BytesSize:    0,
			EntriesCount: 0,
		}
	}

	bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation].MethodsStats[summery.Method].EntriesCount += 1
	bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation].MethodsStats[summery.Method].BytesSize += size
}
