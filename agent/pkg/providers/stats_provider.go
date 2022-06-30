package providers

import (
	"reflect"
	"strings"
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

type TrafficStatsResponse struct {
	PieStats      []*AccumulativeStatsProtocol     `json:"pie"`
	TimelineStats []*AccumulativeStatsProtocolTime `json:"timeline"`
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

func GetGeneralStats() *GeneralStats {
	return &generalStats
}

var protocolMap map[string]*api.Protocol
var protocolToColor map[string]string

func SetProtocolMap(input map[string]*api.Protocol) {
	protocolMap = input
}

func GetTrafficStats() *TrafficStatsResponse {
	bucketsStatsCopy := getBucketStatsCopy()
	interval := calculateInterval(bucketsStatsCopy[0].BucketTime.Unix(), bucketsStatsCopy[len(bucketsStatsCopy)-1].BucketTime.Unix()) // in seconds

	return &TrafficStatsResponse{
		PieStats:      getAccumulativeStats(bucketsStatsCopy),
		TimelineStats: getAccumulativeStatsTiming(bucketsStatsCopy, interval),
	}
}

func calculateInterval(firstTimestamp int64, lastTimestamp int64) time.Duration {
	if time.Duration(lastTimestamp-firstTimestamp)*time.Second < 15*time.Minute {
		return time.Minute
	} else if time.Duration(lastTimestamp-firstTimestamp)*time.Second < time.Hour {
		return time.Minute * 3
	} else if time.Duration(lastTimestamp-firstTimestamp)*time.Second < 7*time.Hour {
		return time.Minute * 30
	} else if time.Duration(lastTimestamp-firstTimestamp)*time.Second < 25*time.Hour {
		return time.Hour * 2
	} else if time.Duration(lastTimestamp-firstTimestamp)*time.Second < 8*24*time.Hour {
		return time.Hour * 12
	}
	return time.Hour * 24
}

func getAccumulativeStats(stats BucketStats) []*AccumulativeStatsProtocol {
	if len(stats) == 0 {
		return make([]*AccumulativeStatsProtocol, 0)
	}

	methodsPerProtocolAggregated := getAggregatedStats(stats)

	return convertAccumulativeStatsDictToArray(methodsPerProtocolAggregated)
}

func getAccumulativeStatsTiming(stats BucketStats, interval time.Duration) []*AccumulativeStatsProtocolTime {
	if len(stats) == 0 {
		return make([]*AccumulativeStatsProtocolTime, 0)
	}

	methodsPerProtocolPerTimeAggregated := getAggregatedResultTiming(interval, stats)

	return convertAccumulativeStatsTimelineDictToArray(methodsPerProtocolPerTimeAggregated)
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

func convertAccumulativeStatsTimelineDictToArray(methodsPerProtocolPerTimeAggregated map[time.Time]map[string]map[string]*AccumulativeStatsCounter) []*AccumulativeStatsProtocolTime {
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
				Color:   getColor(protocolName),
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

func convertAccumulativeStatsDictToArray(methodsPerProtocolAggregated map[string]map[string]*AccumulativeStatsCounter) []*AccumulativeStatsProtocol {
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
			Color:   getColor(protocolName),
			Methods: methods,
		})
	}
	return protocolsData
}

func getColor(protocolName string) string {
	if protocolToColor == nil {
		result := map[string]string{}
		for item, value := range protocolMap {
			result[strings.Split(item, "/")[2]] = value.BackgroundColor
		}
		protocolToColor = result
	}
	return protocolToColor[protocolName]
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

func getAggregatedResultTiming(interval time.Duration, stats BucketStats) map[time.Time]map[string]map[string]*AccumulativeStatsCounter {
	methodsPerProtocolPerTimeAggregated := map[time.Time]map[string]map[string]*AccumulativeStatsCounter{}

	bucketStatsIndex := len(stats) - 1
	for bucketStatsIndex >= 0 {
		currentBucketTime := stats[bucketStatsIndex].BucketTime
		resultBucketRoundedKey := currentBucketTime.Add(-1 * interval / 2).Round(interval)

		for protocolName, data := range stats[bucketStatsIndex].ProtocolStats {
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

		bucketStatsIndex--
	}
	return methodsPerProtocolPerTimeAggregated
}

func getAggregatedStats(bucketStatsCopy BucketStats) map[string]map[string]*AccumulativeStatsCounter {
	methodsPerProtocolAggregated := make(map[string]map[string]*AccumulativeStatsCounter, 0)
	for _, countersOfTimeFrame := range bucketStatsCopy {
		for protocolName, value := range countersOfTimeFrame.ProtocolStats {
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
	return methodsPerProtocolAggregated
}
