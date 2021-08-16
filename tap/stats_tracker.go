package tap

import (
	"sync"
	"time"
)

type AppStats struct {
	StartTime                        time.Time `json:"-"`
	MatchedMessages                  int       `json:"-"`
	TotalProcessedBytes              int64     `json:"totalProcessedBytes"`
	TotalPacketsCount                int64     `json:"totalPacketsCount"`
	TotalTcpPacketsCount             int64     `json:"totalTcpPacketsCount"`
	TotalReassembledTcpPayloadsCount int64     `json:"totalReassembledTcpPayloadsCount"`
	TotalTlsConnectionsCount         int64     `json:"totalTlsConnectionsCount"`
	TotalMatchedMessages             int64     `json:"totalMatchedMessages"`
}

type StatsTracker struct {
	appStats                              AppStats
	matchedMessagesMutex                  sync.Mutex
	totalProcessedSizeMutex               sync.Mutex
	totalPacketsCountMutex                sync.Mutex
	totalTcpPacketsCountMutex             sync.Mutex
	totalReassembledTcpPayloadsCountMutex sync.Mutex
	totalTlsConnectionsCountMutex         sync.Mutex
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

func (st *StatsTracker) incTcpPacketsCount() {
	st.totalTcpPacketsCountMutex.Lock()
	st.appStats.TotalTcpPacketsCount++
	st.totalTcpPacketsCountMutex.Unlock()
}

func (st *StatsTracker) incReassembledTcpPayloadsCount() {
	st.totalReassembledTcpPayloadsCountMutex.Lock()
	st.appStats.TotalReassembledTcpPayloadsCount++
	st.totalReassembledTcpPayloadsCountMutex.Unlock()
}

func (st *StatsTracker) incTlsConnectionsCount() {
	st.totalTlsConnectionsCountMutex.Lock()
	st.appStats.TotalTlsConnectionsCount++
	st.totalTlsConnectionsCountMutex.Unlock()
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
