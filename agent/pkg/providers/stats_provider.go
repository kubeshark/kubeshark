package providers

import (
	"reflect"
	"sync"
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
	generalStats      = GeneralStats{}
	bucketsStats      = BucketStats{}
	bucketStatsLocker = sync.Mutex{}
)

const (
	InternalBucketThreshold = time.Minute * 1
)

func ResetGeneralStats() {
	generalStats = GeneralStats{}
}

func GetGeneralStats() GeneralStats {
	return generalStats
}

func GetAccumulativeStats() []*AccumulativeStatsProtocol {
	bucketStatsCopy := getBucketStatsCopy()
	if len(bucketStatsCopy) == 0 {
		return make([]*AccumulativeStatsProtocol, 0)
	}

	methodsPerProtocolAggregated, protocolToColor := getAggregatedStatsAllTime(bucketStatsCopy)

	return convertAccumulativeStatsDictToArray(methodsPerProtocolAggregated, protocolToColor)
}

func GetAccumulativeStatsTiming(intervalSeconds int, numberOfBars int) []*AccumulativeStatsProtocolTime {
	bucketStatsCopy := getBucketStatsCopy()
	if len(bucketStatsCopy) == 0 {
		return make([]*AccumulativeStatsProtocolTime, 0)
	}

	firstBucketTime := getFirstBucketTime(time.Now().UTC(), intervalSeconds, numberOfBars)

	methodsPerProtocolPerTimeAggregated, protocolToColor := getAggregatedResultTimingFromSpecificTime(intervalSeconds, bucketStatsCopy, firstBucketTime)

	return convertAccumulativeStatsTimelineDictToArray(methodsPerProtocolPerTimeAggregated, protocolToColor)
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
	entryTimeBucketRounded := getBucketFromTimeStamp(summery.Timestamp)

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

func getBucketFromTimeStamp(timestamp int64) time.Time {
	entryTimeStampAsTime := time.UnixMilli(timestamp)
	return entryTimeStampAsTime.Add(-1 * InternalBucketThreshold / 2).Round(InternalBucketThreshold)
}

func getFirstBucketTime(endTime time.Time, intervalSeconds int, numberOfBars int) time.Time {
	lastBucketTime := endTime.Add(-1 * time.Second * time.Duration(intervalSeconds) / 2).Round(time.Second * time.Duration(intervalSeconds))
	firstBucketTime := lastBucketTime.Add(-1 * time.Second * time.Duration(intervalSeconds*(numberOfBars-1)))
	return firstBucketTime
}

func convertAccumulativeStatsTimelineDictToArray(methodsPerProtocolPerTimeAggregated map[time.Time]map[string]map[string]*AccumulativeStatsCounter, protocolToColor map[string]string) []*AccumulativeStatsProtocolTime {
	finalResult := make([]*AccumulativeStatsProtocolTime, 0)
	for timeKey, item := range methodsPerProtocolPerTimeAggregated {
		protocolsData := make([]*AccumulativeStatsProtocol, 0)
		for protocolName := range item {
			entriesCount := 0
			volumeSizeBytes := 0
			methods := make([]*AccumulativeStatsCounter, 0)
			for _, methodAccData := range methodsPerProtocolPerTimeAggregated[timeKey][protocolName] {
				entriesCount += methodAccData.EntriesCount
				volumeSizeBytes += methodAccData.VolumeSizeBytes
				methods = append(methods, methodAccData)
			}
			protocolsData = append(protocolsData, &AccumulativeStatsProtocol{
				AccumulativeStatsCounter: AccumulativeStatsCounter{
					Name:            protocolName,
					EntriesCount:    entriesCount,
					VolumeSizeBytes: volumeSizeBytes,
				},
				Color:   protocolToColor[protocolName],
				Methods: methods,
			})
		}
		finalResult = append(finalResult, &AccumulativeStatsProtocolTime{
			Time:          timeKey.UnixMilli(),
			ProtocolsData: protocolsData,
		})
	}
	return finalResult
}

func convertAccumulativeStatsDictToArray(methodsPerProtocolAggregated map[string]map[string]*AccumulativeStatsCounter, protocolToColor map[string]string) []*AccumulativeStatsProtocol {
	protocolsData := make([]*AccumulativeStatsProtocol, 0)
	for protocolName, value := range methodsPerProtocolAggregated {
		entriesCount := 0
		volumeSizeBytes := 0
		methods := make([]*AccumulativeStatsCounter, 0)
		for _, methodAccData := range value {
			entriesCount += methodAccData.EntriesCount
			volumeSizeBytes += methodAccData.VolumeSizeBytes
			methods = append(methods, methodAccData)
		}
		protocolsData = append(protocolsData, &AccumulativeStatsProtocol{
			AccumulativeStatsCounter: AccumulativeStatsCounter{
				Name:            protocolName,
				EntriesCount:    entriesCount,
				VolumeSizeBytes: volumeSizeBytes,
			},
			Color:   protocolToColor[protocolName],
			Methods: methods,
		})
	}
	return protocolsData
}

func getBucketStatsCopy() BucketStats {
	bucketStatsCopy := BucketStats{}
	bucketStatsLocker.Lock()
	if err := copier.Copy(&bucketStatsCopy, bucketsStats); err != nil {
		logger.Log.Errorf("Error while copying src stats into temporary copied object")
		return nil
	}
	bucketStatsLocker.Unlock()
	return bucketStatsCopy
}

func getAggregatedResultTimingFromSpecificTime(intervalSeconds int, bucketStats BucketStats, firstBucketTime time.Time) (map[time.Time]map[string]map[string]*AccumulativeStatsCounter, map[string]string) {
	protocolToColor := map[string]string{}
	methodsPerProtocolPerTimeAggregated := map[time.Time]map[string]map[string]*AccumulativeStatsCounter{}

	bucketStatsIndex := len(bucketStats) - 1
	for bucketStatsIndex >= 0 {
		currentBucketTime := bucketStats[bucketStatsIndex].BucketTime
		if currentBucketTime.After(firstBucketTime) || currentBucketTime.Equal(firstBucketTime) {
			resultBucketRoundedKey := currentBucketTime.Add(-1 * time.Second * time.Duration(intervalSeconds) / 2).Round(time.Second * time.Duration(intervalSeconds))

			for protocolName, data := range bucketStats[bucketStatsIndex].ProtocolStats {
				if _, ok := protocolToColor[protocolName]; !ok {
					protocolToColor[protocolName] = data.Color
				}

				for methodName, dataOfMethod := range data.MethodsStats {

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
		}
		bucketStatsIndex--
	}
	return methodsPerProtocolPerTimeAggregated, protocolToColor
}

func getAggregatedStatsAllTime(bucketStatsCopy BucketStats) (map[string]map[string]*AccumulativeStatsCounter, map[string]string) {
	protocolToColor := make(map[string]string, 0)
	methodsPerProtocolAggregated := make(map[string]map[string]*AccumulativeStatsCounter, 0)
	for _, countersOfTimeFrame := range bucketStatsCopy {
		for protocolName, value := range countersOfTimeFrame.ProtocolStats {
			if _, ok := protocolToColor[protocolName]; !ok {
				protocolToColor[protocolName] = value.Color
			}

			for method, countersValue := range value.MethodsStats {
				if _, found := methodsPerProtocolAggregated[protocolName]; !found {
					methodsPerProtocolAggregated[protocolName] = map[string]*AccumulativeStatsCounter{}
				}
				if _, found := methodsPerProtocolAggregated[protocolName][method]; !found {
					methodsPerProtocolAggregated[protocolName][method] = &AccumulativeStatsCounter{
						Name:            method,
						EntriesCount:    0,
						VolumeSizeBytes: 0,
					}
				}
				methodsPerProtocolAggregated[protocolName][method].EntriesCount += countersValue.EntriesCount
				methodsPerProtocolAggregated[protocolName][method].VolumeSizeBytes += countersValue.VolumeInBytes
			}
		}
	}
	return methodsPerProtocolAggregated, protocolToColor
}
