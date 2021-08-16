package tap

import (
	"sync"
	"time"
)

type AppStats struct {
	StartTime                   time.Time `json:"-"`
	ProcessedBytes              int64     `json:"processedBytes"`
	PacketsCount                int64     `json:"packetsCount"`
	TcpPacketsCount             int64     `json:"tcpPacketsCount"`
	ReassembledTcpPayloadsCount int64     `json:"reassembledTcpPayloadsCount"`
	TlsConnectionsCount         int64     `json:"tlsConnectionsCount"`
	MatchedMessages             int64     `json:"matchedMessages"`
}

type StatsTracker struct {
	appStats                         AppStats
	ProcessedBytesMutex              sync.Mutex
	PacketsCountMutex                sync.Mutex
	TcpPacketsCountMutex             sync.Mutex
	ReassembledTcpPayloadsCountMutex sync.Mutex
	TlsConnectionsCountMutex         sync.Mutex
	matchedMessagesMutex             sync.Mutex
}

func (st *StatsTracker) incMatchedMessages() {
	st.matchedMessagesMutex.Lock()
	st.appStats.MatchedMessages++
	st.matchedMessagesMutex.Unlock()
}

func (st *StatsTracker) incPacketsCount() int64 {
	st.PacketsCountMutex.Lock()
	st.appStats.PacketsCount++
	currentPacketsCount := st.appStats.PacketsCount
	st.PacketsCountMutex.Unlock()
	return currentPacketsCount
}

func (st *StatsTracker) incTcpPacketsCount() {
	st.TcpPacketsCountMutex.Lock()
	st.appStats.TcpPacketsCount++
	st.TcpPacketsCountMutex.Unlock()
}

func (st *StatsTracker) incReassembledTcpPayloadsCount() {
	st.ReassembledTcpPayloadsCountMutex.Lock()
	st.appStats.ReassembledTcpPayloadsCount++
	st.ReassembledTcpPayloadsCountMutex.Unlock()
}

func (st *StatsTracker) incTlsConnectionsCount() {
	st.TlsConnectionsCountMutex.Lock()
	st.appStats.TlsConnectionsCount++
	st.TlsConnectionsCountMutex.Unlock()
}

func (st *StatsTracker) updateProcessedBytes(size int64) {
	st.ProcessedBytesMutex.Lock()
	st.appStats.ProcessedBytes += size
	st.ProcessedBytesMutex.Unlock()
}

func (st *StatsTracker) setStartTime(startTime time.Time) {
	st.appStats.StartTime = startTime
}

func (st *StatsTracker) dumpStats() *AppStats {
	currentAppStats := &AppStats{StartTime: st.appStats.StartTime}

	st.ProcessedBytesMutex.Lock()
	currentAppStats.ProcessedBytes = st.appStats.ProcessedBytes
	st.appStats.ProcessedBytes = 0
	st.ProcessedBytesMutex.Unlock()

	st.PacketsCountMutex.Lock()
	currentAppStats.PacketsCount = st.appStats.PacketsCount
	st.appStats.PacketsCount = 0
	st.PacketsCountMutex.Unlock()

	st.TcpPacketsCountMutex.Lock()
	currentAppStats.TcpPacketsCount = st.appStats.TcpPacketsCount
	st.appStats.TcpPacketsCount = 0
	st.TcpPacketsCountMutex.Unlock()

	st.ReassembledTcpPayloadsCountMutex.Lock()
	currentAppStats.ReassembledTcpPayloadsCount = st.appStats.ReassembledTcpPayloadsCount
	st.appStats.ReassembledTcpPayloadsCount = 0
	st.ReassembledTcpPayloadsCountMutex.Unlock()

	st.TlsConnectionsCountMutex.Lock()
	currentAppStats.TlsConnectionsCount = st.appStats.TlsConnectionsCount
	st.appStats.TlsConnectionsCount = 0
	st.TlsConnectionsCountMutex.Unlock()

	st.matchedMessagesMutex.Lock()
	currentAppStats.MatchedMessages = st.appStats.MatchedMessages
	st.appStats.MatchedMessages = 0
	st.matchedMessagesMutex.Unlock()

	return currentAppStats
}
