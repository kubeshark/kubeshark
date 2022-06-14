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
	ProtocolCounters    map[string]map[string]*SizeAndRequestCount
}

type SizeAndRequestCount struct {
	Count     int
	BytesSize int
}

type AccumulativeStatsMethod struct {
	MethodName   string `json:"methodName"`
	RequestCount int    `json:"requestCount"`
	ByteCount    int    `json:"byteCount"`
}

type AccumulativeStatsProtocol struct {
	ProtocolName string                     `json:"protocolName"`
	RequestCount int                        `json:"requestCount"`
	ByteCount    int                        `json:"byteCount"`
	Methods      []*AccumulativeStatsMethod `json:"methods"`
}

var generalStats = InitGeneralStats()

func ResetGeneralStats() {
	generalStats = InitGeneralStats()
}

func InitGeneralStats() GeneralStats {
	generalStatsObj := GeneralStats{}
	generalStatsObj.ProtocolCounters = map[string]map[string]*SizeAndRequestCount{}
	return generalStatsObj
}

func GetGeneralStats() GeneralStats {
	return generalStats
}

func GetAccumulativeStats() []*AccumulativeStatsProtocol {
	result := make([]*AccumulativeStatsProtocol, 0)
	for protocolName, value := range generalStats.ProtocolCounters {
		totalProtocolRequestCount := 0
		totalBytesProtocol := 0
		methods := make([]*AccumulativeStatsMethod, 0)

		for method, countersValue := range value {
			methodData := &AccumulativeStatsMethod{
				MethodName:   method,
				RequestCount: 0,
				ByteCount:    0,
			}
			totalProtocolRequestCount += countersValue.Count
			methodData.RequestCount += countersValue.Count
			totalBytesProtocol += countersValue.BytesSize
			methodData.ByteCount += countersValue.BytesSize
			methods = append(methods, methodData)
		}
		newProtocolData := &AccumulativeStatsProtocol{
			ProtocolName: protocolName,
			RequestCount: totalProtocolRequestCount,
			ByteCount:    totalBytesProtocol,
			Methods:      methods,
		}
		result = append(result, newProtocolData)
	}
	return result
}

func EntryAdded(size int, summery api.BaseEntry) {
	generalStats.EntriesCount++
	generalStats.EntriesVolumeInGB += float64(size) / (1 << 30)

	currentTimestamp := int(time.Now().Unix())

	if reflect.Value.IsZero(reflect.ValueOf(generalStats.FirstEntryTimestamp)) {
		generalStats.FirstEntryTimestamp = currentTimestamp
	}

	if _, found := generalStats.ProtocolCounters[summery.Protocol.Name]; !found {
		generalStats.ProtocolCounters[summery.Protocol.Name] = map[string]*SizeAndRequestCount{}
	}
	if _, found := generalStats.ProtocolCounters[summery.Protocol.Name][summery.Method]; !found {
		generalStats.ProtocolCounters[summery.Protocol.Name][summery.Method] = &SizeAndRequestCount{
			BytesSize: 0,
			Count:     0,
		}
	}

	generalStats.ProtocolCounters[summery.Protocol.Name][summery.Method].Count += 1
	generalStats.ProtocolCounters[summery.Protocol.Name][summery.Method].BytesSize += size

	generalStats.LastEntryTimestamp = currentTimestamp
}
