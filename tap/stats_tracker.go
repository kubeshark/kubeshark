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
	processedBytesMutex              sync.Mutex
	packetsCountMutex                sync.Mutex
	tcpPacketsCountMutex             sync.Mutex
	reassembledTcpPayloadsCountMutex sync.Mutex
	tlsConnectionsCountMutex         sync.Mutex
	matchedMessagesMutex             sync.Mutex
}

func (st *StatsTracker) incMatchedMessages() {
	st.matchedMessagesMutex.Lock()
	st.appStats.MatchedMessages++
	st.matchedMessagesMutex.Unlock()
}

func (st *StatsTracker) incPacketsCount() int64 {
	st.packetsCountMutex.Lock()
	st.appStats.PacketsCount++
	currentPacketsCount := st.appStats.PacketsCount
	st.packetsCountMutex.Unlock()
	return currentPacketsCount
}

func (st *StatsTracker) incTcpPacketsCount() {
	st.tcpPacketsCountMutex.Lock()
	st.appStats.TcpPacketsCount++
	st.tcpPacketsCountMutex.Unlock()
}

func (st *StatsTracker) incReassembledTcpPayloadsCount() {
	st.reassembledTcpPayloadsCountMutex.Lock()
	st.appStats.ReassembledTcpPayloadsCount++
	st.reassembledTcpPayloadsCountMutex.Unlock()
}

func (st *StatsTracker) incTlsConnectionsCount() {
	st.tlsConnectionsCountMutex.Lock()
	st.appStats.TlsConnectionsCount++
	st.tlsConnectionsCountMutex.Unlock()
}

func (st *StatsTracker) updateProcessedBytes(size int64) {
	st.processedBytesMutex.Lock()
	st.appStats.ProcessedBytes += size
	st.processedBytesMutex.Unlock()
}

func (st *StatsTracker) setStartTime(startTime time.Time) {
	st.appStats.StartTime = startTime
}

func (st *StatsTracker) dumpStats() *AppStats {
	currentAppStats := &AppStats{StartTime: st.appStats.StartTime}

	st.processedBytesMutex.Lock()
	currentAppStats.ProcessedBytes = st.appStats.ProcessedBytes
	st.appStats.ProcessedBytes = 0
	st.processedBytesMutex.Unlock()

	st.packetsCountMutex.Lock()
	currentAppStats.PacketsCount = st.appStats.PacketsCount
	st.appStats.PacketsCount = 0
	st.packetsCountMutex.Unlock()

	st.tcpPacketsCountMutex.Lock()
	currentAppStats.TcpPacketsCount = st.appStats.TcpPacketsCount
	st.appStats.TcpPacketsCount = 0
	st.tcpPacketsCountMutex.Unlock()

	st.reassembledTcpPayloadsCountMutex.Lock()
	currentAppStats.ReassembledTcpPayloadsCount = st.appStats.ReassembledTcpPayloadsCount
	st.appStats.ReassembledTcpPayloadsCount = 0
	st.reassembledTcpPayloadsCountMutex.Unlock()

	st.tlsConnectionsCountMutex.Lock()
	currentAppStats.TlsConnectionsCount = st.appStats.TlsConnectionsCount
	st.appStats.TlsConnectionsCount = 0
	st.tlsConnectionsCountMutex.Unlock()

	st.matchedMessagesMutex.Lock()
	currentAppStats.MatchedMessages = st.appStats.MatchedMessages
	st.appStats.MatchedMessages = 0
	st.matchedMessagesMutex.Unlock()

	return currentAppStats
}
