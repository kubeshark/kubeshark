package api

import (
	"sync/atomic"
	"time"
)

type AppStats struct {
	StartTime                   time.Time `json:"-"`
	ProcessedBytes              uint64    `json:"processedBytes"`
	PacketsCount                uint64    `json:"packetsCount"`
	TcpPacketsCount             uint64    `json:"tcpPacketsCount"`
	TapperPacketsCount          uint64    `json:"tapperPacketsCount"`
	ReassembledTcpPayloadsCount uint64    `json:"reassembledTcpPayloadsCount"`
	TlsConnectionsCount         uint64    `json:"tlsConnectionsCount"`
	MatchedPairs                uint64    `json:"matchedPairs"`
	DroppedTcpStreams           uint64    `json:"droppedTcpStreams"`
}

func (as *AppStats) IncMatchedPairs() {
	atomic.AddUint64(&as.MatchedPairs, 1)
}

func (as *AppStats) IncDroppedTcpStreams() {
	atomic.AddUint64(&as.DroppedTcpStreams, 1)
}

func (as *AppStats) IncPacketsCount() uint64 {
	atomic.AddUint64(&as.PacketsCount, 1)
	return as.PacketsCount
}

func (as *AppStats) IncTcpPacketsCount() {
	atomic.AddUint64(&as.TcpPacketsCount, 1)
}

func (as *AppStats) IncTapperPacketsCount() {
	atomic.AddUint64(&as.TapperPacketsCount, 1)
}

func (as *AppStats) IncReassembledTcpPayloadsCount() {
	atomic.AddUint64(&as.ReassembledTcpPayloadsCount, 1)
}

func (as *AppStats) IncTlsConnectionsCount() {
	atomic.AddUint64(&as.TlsConnectionsCount, 1)
}

func (as *AppStats) UpdateProcessedBytes(size uint64) {
	atomic.AddUint64(&as.ProcessedBytes, size)
}

func (as *AppStats) SetStartTime(startTime time.Time) {
	as.StartTime = startTime
}

func (as *AppStats) DumpStats() *AppStats {
	currentAppStats := &AppStats{StartTime: as.StartTime}

	currentAppStats.ProcessedBytes = resetUint64(&as.ProcessedBytes)
	currentAppStats.PacketsCount = resetUint64(&as.PacketsCount)
	currentAppStats.TcpPacketsCount = resetUint64(&as.TcpPacketsCount)
	currentAppStats.TapperPacketsCount = resetUint64(&as.TapperPacketsCount)
	currentAppStats.ReassembledTcpPayloadsCount = resetUint64(&as.ReassembledTcpPayloadsCount)
	currentAppStats.TlsConnectionsCount = resetUint64(&as.TlsConnectionsCount)
	currentAppStats.MatchedPairs = resetUint64(&as.MatchedPairs)
	currentAppStats.DroppedTcpStreams = resetUint64(&as.DroppedTcpStreams)

	return currentAppStats
}

func resetUint64(ref *uint64) (val uint64) {
	val = atomic.LoadUint64(ref)
	atomic.StoreUint64(ref, 0)
	return
}
