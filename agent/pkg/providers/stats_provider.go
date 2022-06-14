package providers

import (
	"reflect"
	"time"

	"github.com/up9inc/mizu/agent/pkg/utils"
	"github.com/up9inc/mizu/tap/api"
)

type GeneralStats struct {
	EntriesCount           int
	EntriesVolumeInGB      float64
	FirstEntryTimestamp    int
	LastEntryTimestamp     int
	CountPerProtocolMethod map[string]map[string]int
	SizePerProtocolMethod  map[string]map[string]int
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
	generalStatsObj.CountPerProtocolMethod = map[string]map[string]int{}
	generalStatsObj.SizePerProtocolMethod = map[string]map[string]int{}
	return generalStatsObj
}

func GetGeneralStats() GeneralStats {
	return generalStats
}

func GetAccumulativeStats() []*AccumulativeStatsProtocol {
	allProtocols := make([]string, 0)
	for protocolName, _ := range generalStats.CountPerProtocolMethod {
		allProtocols = append(allProtocols, protocolName)
	}
	for protocolName, _ := range generalStats.SizePerProtocolMethod {
		allProtocols = append(allProtocols, protocolName)
	}
	allProtocols = utils.UniqueStringSlice(allProtocols)

	result := make([]*AccumulativeStatsProtocol, 0)
	for _, protocol := range allProtocols {
		totalProtocolRequestCount := 0
		totalBytesProtocol := 0
		methods := make([]*AccumulativeStatsMethod, 0)

		allProtocolMethods := make([]string, 0)
		for protocolName, _ := range generalStats.CountPerProtocolMethod {
			allProtocolMethods = append(allProtocolMethods, protocolName)
		}
		for protocolName, _ := range generalStats.SizePerProtocolMethod {
			allProtocolMethods = append(allProtocolMethods, protocolName)
		}
		allProtocolMethods = utils.UniqueStringSlice(allProtocolMethods)

		for _, method := range allProtocolMethods {
			methodData := &AccumulativeStatsMethod{
				MethodName:   method,
				RequestCount: 0,
				ByteCount:    0,
			}
			if value, ok := generalStats.CountPerProtocolMethod[protocol][method]; ok {
				totalProtocolRequestCount += value
				methodData.RequestCount += value
			}
			if value, ok := generalStats.SizePerProtocolMethod[protocol][method]; ok {
				totalBytesProtocol += value
				methodData.ByteCount += value
			}
			methods = append(methods, methodData)
		}
		newProtocolData := &AccumulativeStatsProtocol{
			ProtocolName: protocol,
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

	if _, found := generalStats.CountPerProtocolMethod[summery.Protocol.Name]; !found {
		generalStats.CountPerProtocolMethod[summery.Protocol.Name] = map[string]int{}
	}
	if _, found := generalStats.CountPerProtocolMethod[summery.Protocol.Name][summery.Method]; !found {
		generalStats.CountPerProtocolMethod[summery.Protocol.Name][summery.Method] = 0
	}

	if _, found := generalStats.SizePerProtocolMethod[summery.Protocol.Name]; !found {
		generalStats.SizePerProtocolMethod[summery.Protocol.Name] = map[string]int{}
	}
	if _, found := generalStats.SizePerProtocolMethod[summery.Protocol.Name][summery.Method]; !found {
		generalStats.SizePerProtocolMethod[summery.Protocol.Name][summery.Method] = 0
	}
	generalStats.CountPerProtocolMethod[summery.Protocol.Name][summery.Method] += 1
	generalStats.SizePerProtocolMethod[summery.Protocol.Name][summery.Method] += size

	generalStats.LastEntryTimestamp = currentTimestamp
}
