package providers

import (
	"reflect"
	"time"

	"github.com/jinzhu/copier"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
)

type GeneralStats struct {
	EntriesCount        int
	EntriesVolumeInGB   float64
	FirstEntryTimestamp int
	LastEntryTimestamp  int
}

type BucketStats []*TimeFrameStatsValue

type TimeFrameStatsValue struct {
	BucketTime    time.Time                `json:"timestamp"`
	ProtocolStats map[string]ProtocolStats `json:"protocols"`
}

type ProtocolStats struct {
	MethodsStats map[string]*SizeAndEntriesCount `json:"methods"`
	Color        string                          `json:"color"`
}

type SizeAndEntriesCount struct {
	EntriesCount  int `json:"entriesCount"`
	VolumeInBytes int `json:"volumeInBytes"`
}

type AccumulativeStatsCounter struct {
	Name            string `json:"name"`
	EntriesCount    int    `json:"entriesCount"`
	VolumeSizeBytes int    `json:"volumeSizeBytes"`
}

type AccumulativeStatsProtocol struct {
	AccumulativeStatsCounter
	Color   string                      `json:"color"`
	Methods []*AccumulativeStatsCounter `json:"methods"`
}

type AccumulativeStatsProtocolTime struct {
	ProtocolsData []*AccumulativeStatsProtocol `json:"protocols"`
	Time          int64                        `json:"timestamp"`
}

var (
	generalStats            = GeneralStats{}
	bucketsStats            = BucketStats{}
	internalBucketThreshold = time.Minute * 1
)

func ResetGeneralStats() {
	generalStats = GeneralStats{}
}

func GetGeneralStats() GeneralStats {
	return generalStats
}

func GetAccumulativeStats() []*AccumulativeStatsProtocol {
	bucketStatsCopy := BucketStats{}
	if err := copier.Copy(&bucketStatsCopy, bucketsStats); err != nil {
		logger.Log.Errorf("Error while copying src stats into temporary copied object")
		return make([]*AccumulativeStatsProtocol, 0)
	}

	result := make(map[string]*AccumulativeStatsProtocol, 0)
	methodsPerProtocolAggregated := make(map[string]map[string]*AccumulativeStatsCounter, 0)
	for _, countersOfTimeFrame := range bucketStatsCopy {
		for protocolName, value := range countersOfTimeFrame.ProtocolStats {

			if _, found := result[protocolName]; !found {
				result[protocolName] = &AccumulativeStatsProtocol{
					AccumulativeStatsCounter: AccumulativeStatsCounter{
						Name:            protocolName,
						EntriesCount:    0,
						VolumeSizeBytes: 0,
					},
					Color: value.Color,
				}
			}
			if _, found := methodsPerProtocolAggregated[protocolName]; !found {
				methodsPerProtocolAggregated[protocolName] = map[string]*AccumulativeStatsCounter{}
			}

			for method, countersValue := range value.MethodsStats {
				if _, found := methodsPerProtocolAggregated[protocolName][method]; !found {
					methodsPerProtocolAggregated[protocolName][method] = &AccumulativeStatsCounter{
						Name:            method,
						EntriesCount:    0,
						VolumeSizeBytes: 0,
					}
				}

				result[protocolName].AccumulativeStatsCounter.EntriesCount += countersValue.EntriesCount
				methodsPerProtocolAggregated[protocolName][method].EntriesCount += countersValue.EntriesCount
				result[protocolName].AccumulativeStatsCounter.VolumeSizeBytes += countersValue.VolumeInBytes
				methodsPerProtocolAggregated[protocolName][method].VolumeSizeBytes += countersValue.VolumeInBytes
			}
		}
	}

	finalResult := make([]*AccumulativeStatsProtocol, 0)
	for _, value := range result {
		methodsForProtocol := make([]*AccumulativeStatsCounter, 0)
		for _, methodValue := range methodsPerProtocolAggregated[value.Name] {
			methodsForProtocol = append(methodsForProtocol, methodValue)
		}
		value.Methods = methodsForProtocol
		finalResult = append(finalResult, value)
	}
	return finalResult
}

func GetAccumulativeStatsTiming(intervalSeconds int, numberOfBars int) []*AccumulativeStatsProtocolTime {
	bucketStatsCopy := BucketStats{}
	if err := copier.Copy(&bucketStatsCopy, bucketsStats); err != nil {
		logger.Log.Errorf("Error while copying src stats into temporary copied object")
		return make([]*AccumulativeStatsProtocolTime, 0)
	}

	if len(bucketStatsCopy) == 0 {
		return make([]*AccumulativeStatsProtocolTime, 0)
	}

	result := make(map[time.Time]map[string]*AccumulativeStatsProtocol, 0)
	methodsPerProtocolPerTimeAggregated := make(map[time.Time]map[string]map[string]*AccumulativeStatsCounter, 0)
	lastBucketTime := time.Now().UTC().Add(-1 * internalBucketThreshold / 2).Round(internalBucketThreshold)
	firstBucketTime := lastBucketTime.Add(-1 * time.Second * time.Duration(intervalSeconds*numberOfBars))
	bucketStatsIndex := len(bucketStatsCopy) - 1

	for bucketStatsIndex >= 0 && (bucketStatsCopy[bucketStatsIndex].BucketTime.Before(lastBucketTime) || bucketStatsCopy[bucketStatsIndex].BucketTime.Equal(lastBucketTime)) &&
		(bucketStatsCopy[bucketStatsIndex].BucketTime.After(firstBucketTime) || bucketStatsCopy[bucketStatsIndex].BucketTime.Equal(firstBucketTime)) {

		resultBucketRoundedKey := bucketStatsCopy[bucketStatsIndex].BucketTime.Round(time.Second * time.Duration(intervalSeconds))

		if _, ok := result[resultBucketRoundedKey]; !ok {
			result[resultBucketRoundedKey] = map[string]*AccumulativeStatsProtocol{}
		}

		for protocolName, data := range bucketStatsCopy[bucketStatsIndex].ProtocolStats {
			if _, ok := result[resultBucketRoundedKey][protocolName]; !ok {
				result[resultBucketRoundedKey][protocolName] = &AccumulativeStatsProtocol{
					AccumulativeStatsCounter: AccumulativeStatsCounter{
						Name:            protocolName,
						EntriesCount:    0,
						VolumeSizeBytes: 0,
					},
					Color:   data.Color,
					Methods: make([]*AccumulativeStatsCounter, 0),
				}
			}

			for methodName, dataOfMethod := range data.MethodsStats {
				result[resultBucketRoundedKey][protocolName].EntriesCount += dataOfMethod.EntriesCount
				result[resultBucketRoundedKey][protocolName].VolumeSizeBytes += dataOfMethod.VolumeInBytes

				if _, ok := methodsPerProtocolPerTimeAggregated[resultBucketRoundedKey]; !ok {
					methodsPerProtocolPerTimeAggregated[resultBucketRoundedKey] = map[string]map[string]*AccumulativeStatsCounter{}
				}
				if _, ok := methodsPerProtocolPerTimeAggregated[resultBucketRoundedKey][protocolName]; !ok {
					methodsPerProtocolPerTimeAggregated[resultBucketRoundedKey][protocolName] = map[string]*AccumulativeStatsCounter{}
				}
				if _, ok := methodsPerProtocolPerTimeAggregated[resultBucketRoundedKey][protocolName][methodName]; !ok {
					methodsPerProtocolPerTimeAggregated[resultBucketRoundedKey][protocolName][methodName] = &AccumulativeStatsCounter{
						Name:            methodName,
						EntriesCount:    0,
						VolumeSizeBytes: 0,
					}
				}
				methodsPerProtocolPerTimeAggregated[resultBucketRoundedKey][protocolName][methodName].EntriesCount += dataOfMethod.EntriesCount
				methodsPerProtocolPerTimeAggregated[resultBucketRoundedKey][protocolName][methodName].VolumeSizeBytes += dataOfMethod.VolumeInBytes
			}
		}

		bucketStatsIndex--
	}

	for timeKey, item := range result {
		for protocolName, _ := range item {
			methods := make([]*AccumulativeStatsCounter, 0)
			for _, methodAccData := range methodsPerProtocolPerTimeAggregated[timeKey][protocolName] {
				methods = append(methods, methodAccData)
			}
			result[timeKey][protocolName].Methods = methods
		}
	}

	finalResult := make([]*AccumulativeStatsProtocolTime, 0)
	for timeKey, dataOfBucket := range result {
		protocolStats := make([]*AccumulativeStatsProtocol, 0)
		for _, data := range dataOfBucket {
			protocolStats = append(protocolStats, data)
		}
		value := &AccumulativeStatsProtocolTime{
			ProtocolsData: protocolStats,
			Time:          timeKey.UnixMilli(),
		}
		finalResult = append(finalResult, value)
	}
	return finalResult

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
	entryTimeBucketRounded := time.UnixMilli(summery.Timestamp).Add(-1 * internalBucketThreshold / 2).Round(internalBucketThreshold)
	if len(bucketsStats) == 0 {
		bucketsStats = append(bucketsStats, &TimeFrameStatsValue{
			BucketTime:    entryTimeBucketRounded,
			ProtocolStats: map[string]ProtocolStats{},
		})
	}
	bucketOfEntry := bucketsStats[len(bucketsStats)-1]
	if bucketOfEntry.BucketTime != entryTimeBucketRounded {
		bucketOfEntry = &TimeFrameStatsValue{
			BucketTime:    entryTimeBucketRounded,
			ProtocolStats: map[string]ProtocolStats{},
		}
		bucketsStats = append(bucketsStats, bucketOfEntry)
	}
	if _, found := bucketOfEntry.ProtocolStats[summery.Protocol.Abbreviation]; !found {
		bucketOfEntry.ProtocolStats[summery.Protocol.Abbreviation] = ProtocolStats{
			MethodsStats: map[string]*SizeAndEntriesCount{},
			Color:        summery.Protocol.BackgroundColor,
		}
	}
	if _, found := bucketOfEntry.ProtocolStats[summery.Protocol.Abbreviation].MethodsStats[summery.Method]; !found {
		bucketOfEntry.ProtocolStats[summery.Protocol.Abbreviation].MethodsStats[summery.Method] = &SizeAndEntriesCount{
			VolumeInBytes: 0,
			EntriesCount:  0,
		}
	}

	bucketOfEntry.ProtocolStats[summery.Protocol.Abbreviation].MethodsStats[summery.Method].EntriesCount += 1
	bucketOfEntry.ProtocolStats[summery.Protocol.Abbreviation].MethodsStats[summery.Method].VolumeInBytes += size
}
