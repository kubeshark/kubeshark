package providers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
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
}

type SizeAndEntriesCount struct {
	EntriesCount  int `json:"entriesCount"`
	VolumeInBytes int `json:"volumeInBytes"`
}

type AccumulativeStatsCounter struct {
	Name            string `json:"name"`
	Color           string `json:"color"`
	EntriesCount    int    `json:"entriesCount"`
	VolumeSizeBytes int    `json:"volumeSizeBytes"`
}

type AccumulativeStatsProtocol struct {
	AccumulativeStatsCounter
	Methods []*AccumulativeStatsCounter `json:"methods"`
}

type AccumulativeStatsProtocolTime struct {
	ProtocolsData []*AccumulativeStatsProtocol `json:"protocols"`
	Time          int64                        `json:"timestamp"`
}

type TrafficStatsResponse struct {
	Protocols     []string                         `json:"protocols"`
	PieStats      []*AccumulativeStatsProtocol     `json:"pie"`
	TimelineStats []*AccumulativeStatsProtocolTime `json:"timeline"`
}

var (
	generalStats      = GeneralStats{}
	bucketsStats      = BucketStats{}
	bucketStatsLocker = sync.Mutex{}
	protocolToColor   = map[string]string{}
)

const (
	InternalBucketThreshold = time.Minute * 1
	MaxNumberOfBars         = 30
)

func ResetGeneralStats() {
	generalStats = GeneralStats{}
}

func GetGeneralStats() *GeneralStats {
	return &generalStats
}

func InitProtocolToColor(protocolMap map[string]*api.Protocol) {
	for item, value := range protocolMap {
		protocolToColor[api.GetProtocolSummary(item).Abbreviation] = value.BackgroundColor
	}
}

func GetTrafficStats() *TrafficStatsResponse {
	bucketsStatsCopy := getBucketStatsCopy()

	return &TrafficStatsResponse{
		Protocols:     getAvailableProtocols(bucketsStatsCopy),
		PieStats:      getAccumulativeStats(bucketsStatsCopy),
		TimelineStats: getAccumulativeStatsTiming(bucketsStatsCopy),
	}
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

func calculateInterval(firstTimestamp int64, lastTimestamp int64) time.Duration {
	validDurations := []time.Duration{
		time.Minute,
		time.Minute * 2,
		time.Minute * 3,
		time.Minute * 5,
		time.Minute * 10,
		time.Minute * 15,
		time.Minute * 20,
		time.Minute * 30,
		time.Minute * 45,
		time.Minute * 60,
		time.Minute * 75,
		time.Minute * 90,   // 1.5 minutes
		time.Minute * 120,  // 2 hours
		time.Minute * 150,  // 2.5 hours
		time.Minute * 180,  // 3 hours
		time.Minute * 240,  // 4 hours
		time.Minute * 300,  // 5 hours
		time.Minute * 360,  // 6 hours
		time.Minute * 420,  // 7 hours
		time.Minute * 480,  // 8 hours
		time.Minute * 540,  // 9 hours
		time.Minute * 600,  // 10 hours
		time.Minute * 660,  // 11 hours
		time.Minute * 720,  // 12 hours
		time.Minute * 1440, // 24 hours
	}
	duration := time.Duration(lastTimestamp-firstTimestamp) * time.Second / time.Duration(MaxNumberOfBars)
	for _, validDuration := range validDurations {
		if validDuration-duration >= 0 {
			return validDuration
		}
	}
	return duration.Round(validDurations[len(validDurations)-1])

}

func getAccumulativeStats(stats BucketStats) []*AccumulativeStatsProtocol {
	if len(stats) == 0 {
		return make([]*AccumulativeStatsProtocol, 0)
	}

	methodsPerProtocolAggregated := getAggregatedStats(stats)

	return convertAccumulativeStatsDictToArray(methodsPerProtocolAggregated)
}

func getAccumulativeStatsTiming(stats BucketStats) []*AccumulativeStatsProtocolTime {
	if len(stats) == 0 {
		return make([]*AccumulativeStatsProtocolTime, 0)
	}

	interval := calculateInterval(stats[0].BucketTime.Unix(), stats[len(stats)-1].BucketTime.Unix()) // in seconds
	methodsPerProtocolPerTimeAggregated := getAggregatedResultTiming(stats, interval)

	return convertAccumulativeStatsTimelineDictToArray(methodsPerProtocolPerTimeAggregated)
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
		for protocolName, value := range item {
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
					Color:           protocolToColor[protocolName],
					EntriesCount:    entriesCount,
					VolumeSizeBytes: volumeSizeBytes,
				},
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
				Color:           protocolToColor[protocolName],
				EntriesCount:    entriesCount,
				VolumeSizeBytes: volumeSizeBytes,
			},
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

func getAggregatedResultTiming(stats BucketStats, interval time.Duration) map[time.Time]map[string]map[string]*AccumulativeStatsCounter {
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
						Color:           getColorForMethod(protocolName, methodName),
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

func getAggregatedStats(stats BucketStats) map[string]map[string]*AccumulativeStatsCounter {
	methodsPerProtocolAggregated := make(map[string]map[string]*AccumulativeStatsCounter, 0)
	for _, countersOfTimeFrame := range stats {
		for protocolName, value := range countersOfTimeFrame.ProtocolStats {
			for method, countersValue := range value.MethodsStats {
				if _, found := methodsPerProtocolAggregated[protocolName]; !found {
					methodsPerProtocolAggregated[protocolName] = map[string]*AccumulativeStatsCounter{}
				}
				if _, found := methodsPerProtocolAggregated[protocolName][method]; !found {
					methodsPerProtocolAggregated[protocolName][method] = &AccumulativeStatsCounter{
						Name:            method,
						Color:           getColorForMethod(protocolName, method),
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

func getColorForMethod(protocolName string, methodName string) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%v_%v", protocolName, methodName)))
	input := hex.EncodeToString(hash[:])
	return fmt.Sprintf("#%v", input[:6])
}

func getAvailableProtocols(stats BucketStats) []string {
	protocols := map[string]bool{}
	for _, countersOfTimeFrame := range stats {
		for protocolName := range countersOfTimeFrame.ProtocolStats {
			protocols[protocolName] = true
		}
	}

	result := make([]string, 0)
	for protocol := range protocols {
		result = append(result, protocol)
	}
	result = append(result, "ALL")
	return result
}
