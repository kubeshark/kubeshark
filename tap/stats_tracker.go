package tap

import (
	"sync"
	"time"
)

type AppStats struct {
	MatchedMessages      int       `json:"matchedMessages"`
	PacketsCount         int       `json:"packetsCount"`
	ProcessedBytes       int64     `json:"processedBytes"`
	StartTime            time.Time `json:"startTime"`
	TotalMatchedMessages int       `json:"totalMatchedMessages"`
}

type StatsTracker struct {
	appStats      AppStats
	appStatsMutex sync.Mutex
}

func (st *StatsTracker) incMatchedMessages() {
	st.appStatsMutex.Lock()
	st.appStats.MatchedMessages++
	st.appStatsMutex.Unlock()
}

func (st *StatsTracker) incPacketsCount() int {
	st.appStatsMutex.Lock()
	st.appStats.PacketsCount++
	currentPacketsCount := st.appStats.PacketsCount
	st.appStatsMutex.Unlock()
	return currentPacketsCount
}

func (st *StatsTracker) updateProcessedSize(size int64) {
	st.appStatsMutex.Lock()
	st.appStats.ProcessedBytes += size
	st.appStatsMutex.Unlock()
}

func (st *StatsTracker) setStartTime(startTime time.Time) {
	st.appStats.StartTime = startTime
}

func (st *StatsTracker) dumpStats() int {
	st.appStatsMutex.Lock()
	matchedMessages := st.appStats.MatchedMessages
	st.appStats.TotalMatchedMessages += matchedMessages
	st.appStats.MatchedMessages = 0
	st.appStatsMutex.Unlock()
	return matchedMessages
}

func (st *StatsTracker) GetStats() AppStats {
	return st.appStats
}
