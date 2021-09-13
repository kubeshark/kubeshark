package api

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
	MatchedPairs                int64     `json:"matchedPairs"`
	DroppedTcpStreams           int64     `json:"droppedTcpStreams"`
}

type StatsTracker struct {
	AppStats                         AppStats
	ProcessedBytesMutex              sync.Mutex
	PacketsCountMutex                sync.Mutex
	TcpPacketsCountMutex             sync.Mutex
	ReassembledTcpPayloadsCountMutex sync.Mutex
	TlsConnectionsCountMutex         sync.Mutex
	MatchedPairsMutex                sync.Mutex
	DroppedTcpStreamsMutex           sync.Mutex
}

func (st *StatsTracker) IncMatchedPairs() {
	st.MatchedPairsMutex.Lock()
	st.AppStats.MatchedPairs++
	st.MatchedPairsMutex.Unlock()
}

func (st *StatsTracker) IncDroppedTcpStreams() {
	st.DroppedTcpStreamsMutex.Lock()
	st.AppStats.DroppedTcpStreams++
	st.DroppedTcpStreamsMutex.Unlock()
}

func (st *StatsTracker) IncPacketsCount() int64 {
	st.PacketsCountMutex.Lock()
	st.AppStats.PacketsCount++
	currentPacketsCount := st.AppStats.PacketsCount
	st.PacketsCountMutex.Unlock()
	return currentPacketsCount
}

func (st *StatsTracker) IncTcpPacketsCount() {
	st.TcpPacketsCountMutex.Lock()
	st.AppStats.TcpPacketsCount++
	st.TcpPacketsCountMutex.Unlock()
}

func (st *StatsTracker) IncReassembledTcpPayloadsCount() {
	st.ReassembledTcpPayloadsCountMutex.Lock()
	st.AppStats.ReassembledTcpPayloadsCount++
	st.ReassembledTcpPayloadsCountMutex.Unlock()
}

func (st *StatsTracker) IncTlsConnectionsCount() {
	st.TlsConnectionsCountMutex.Lock()
	st.AppStats.TlsConnectionsCount++
	st.TlsConnectionsCountMutex.Unlock()
}

func (st *StatsTracker) UpdateProcessedBytes(size int64) {
	st.ProcessedBytesMutex.Lock()
	st.AppStats.ProcessedBytes += size
	st.ProcessedBytesMutex.Unlock()
}

func (st *StatsTracker) SetStartTime(startTime time.Time) {
	st.AppStats.StartTime = startTime
}

func (st *StatsTracker) DumpStats() *AppStats {
	currentAppStats := &AppStats{StartTime: st.AppStats.StartTime}

	st.ProcessedBytesMutex.Lock()
	currentAppStats.ProcessedBytes = st.AppStats.ProcessedBytes
	st.AppStats.ProcessedBytes = 0
	st.ProcessedBytesMutex.Unlock()

	st.PacketsCountMutex.Lock()
	currentAppStats.PacketsCount = st.AppStats.PacketsCount
	st.AppStats.PacketsCount = 0
	st.PacketsCountMutex.Unlock()

	st.TcpPacketsCountMutex.Lock()
	currentAppStats.TcpPacketsCount = st.AppStats.TcpPacketsCount
	st.AppStats.TcpPacketsCount = 0
	st.TcpPacketsCountMutex.Unlock()

	st.ReassembledTcpPayloadsCountMutex.Lock()
	currentAppStats.ReassembledTcpPayloadsCount = st.AppStats.ReassembledTcpPayloadsCount
	st.AppStats.ReassembledTcpPayloadsCount = 0
	st.ReassembledTcpPayloadsCountMutex.Unlock()

	st.TlsConnectionsCountMutex.Lock()
	currentAppStats.TlsConnectionsCount = st.AppStats.TlsConnectionsCount
	st.AppStats.TlsConnectionsCount = 0
	st.TlsConnectionsCountMutex.Unlock()

	st.MatchedPairsMutex.Lock()
	currentAppStats.MatchedPairs = st.AppStats.MatchedPairs
	st.AppStats.MatchedPairs = 0
	st.MatchedPairsMutex.Unlock()

	st.DroppedTcpStreamsMutex.Lock()
	currentAppStats.DroppedTcpStreams = st.AppStats.DroppedTcpStreams
	st.AppStats.DroppedTcpStreams = 0
	st.DroppedTcpStreamsMutex.Unlock()

	return currentAppStats
}
