package tap

import (
	"sync"
	"time"
)

type AppStats struct {
	StartTime                        time.Time `json:"-"`
	TotalProcessedBytes              int64     `json:"totalProcessedBytes"`
	TotalPacketsCount                int64     `json:"totalPacketsCount"`
	TotalTcpPacketsCount             int64     `json:"totalTcpPacketsCount"`
	TotalReassembledTcpPayloadsCount int64     `json:"totalReassembledTcpPayloadsCount"`
	TotalTlsConnectionsCount         int64     `json:"totalTlsConnectionsCount"`
	MatchedMessages                  int       `json:"-"`
	TotalMatchedMessages             int64     `json:"totalMatchedMessages"`
}

type StatsTracker struct {
	appStats                              AppStats
	totalProcessedSizeMutex               sync.Mutex
	totalPacketsCountMutex                sync.Mutex
	totalTcpPacketsCountMutex             sync.Mutex
	totalReassembledTcpPayloadsCountMutex sync.Mutex
	totalTlsConnectionsCountMutex         sync.Mutex
	matchedMessagesMutex                  sync.Mutex
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

func (st *StatsTracker) dumpStats() *AppStats {
	currentAppStats := &AppStats{StartTime: st.appStats.StartTime}

	st.totalProcessedSizeMutex.Lock()
	currentAppStats.TotalProcessedBytes = st.appStats.TotalProcessedBytes
	st.appStats.TotalProcessedBytes = 0
	st.totalProcessedSizeMutex.Unlock()

	st.totalPacketsCountMutex.Lock()
	currentAppStats.TotalPacketsCount = st.appStats.TotalPacketsCount
	st.appStats.TotalPacketsCount = 0
	st.totalPacketsCountMutex.Unlock()

	st.totalTcpPacketsCountMutex.Lock()
	currentAppStats.TotalTcpPacketsCount = st.appStats.TotalTcpPacketsCount
	st.appStats.TotalTcpPacketsCount = 0
	st.totalTcpPacketsCountMutex.Unlock()

	st.totalReassembledTcpPayloadsCountMutex.Lock()
	currentAppStats.TotalReassembledTcpPayloadsCount = st.appStats.TotalReassembledTcpPayloadsCount
	st.appStats.TotalReassembledTcpPayloadsCount = 0
	st.totalReassembledTcpPayloadsCountMutex.Unlock()

	st.totalTlsConnectionsCountMutex.Lock()
	currentAppStats.TotalTlsConnectionsCount = st.appStats.TotalTlsConnectionsCount
	st.appStats.TotalTlsConnectionsCount = 0
	st.totalTlsConnectionsCountMutex.Unlock()

	st.matchedMessagesMutex.Lock()
	currentAppStats.MatchedMessages = st.appStats.MatchedMessages
	currentAppStats.TotalMatchedMessages = st.appStats.TotalMatchedMessages
	st.appStats.MatchedMessages = 0
	st.appStats.TotalMatchedMessages = 0
	st.matchedMessagesMutex.Unlock()

	return currentAppStats
}
