package tap

import (
	"sync"
	"time"
)

type AppStats struct {
	StartTime            time.Time `json:"startTime"`
	MatchedMessages      int       `json:"matchedMessages"`
	TotalPacketsCount    int64     `json:"totalPacketsCount"`
	TotalProcessedBytes  int64     `json:"totalProcessedBytes"`
	TotalMatchedMessages int64     `json:"totalMatchedMessages"`
}

type StatsTracker struct {
	appStats                AppStats
	matchedMessagesMutex    sync.Mutex
	totalPacketsCountMutex  sync.Mutex
	totalProcessedSizeMutex sync.Mutex
}

func (st *StatsTracker) incMatchedMessages() {
	st.matchedMessagesMutex.Lock()
	st.appStats.MatchedMessages++
	st.appStats.TotalMatchedMessages++
	st.matchedMessagesMutex.Unlock()
}

func (st *StatsTracker) incPacketsCount() int64 {
	st.totalPacketsCountMutex.Lock()
	st.appStats.TotalPacketsCount++
	currentPacketsCount := st.appStats.TotalPacketsCount
	st.totalPacketsCountMutex.Unlock()
	return currentPacketsCount
}

func (st *StatsTracker) updateProcessedSize(size int64) {
	st.totalProcessedSizeMutex.Lock()
	st.appStats.TotalProcessedBytes += size
	st.totalProcessedSizeMutex.Unlock()
}

func (st *StatsTracker) setStartTime(startTime time.Time) {
	st.appStats.StartTime = startTime
}

func (st *StatsTracker) dumpStats() int {
	st.matchedMessagesMutex.Lock()
	matchedMessages := st.appStats.MatchedMessages
	st.appStats.MatchedMessages = 0
	st.matchedMessagesMutex.Unlock()

	return matchedMessages
}
