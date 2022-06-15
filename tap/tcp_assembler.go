package tap

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/dbgctl"
	"github.com/up9inc/mizu/tap/diagnose"
	"github.com/up9inc/mizu/tap/source"
)

const (
	lastClosedConnectionsMaxItems = 1000
	packetsSeenLogThreshold       = 1000
	lastAckThreshold              = time.Duration(3) * time.Second
)

type AssemblerStats struct {
	flushedConnections int
	closedConnections  int
}

type tcpAssembler struct {
	*reassembly.Assembler
	streamPool             *reassembly.StreamPool
	streamFactory          *tcpStreamFactory
	ignoredPorts           []uint16
	lastClosedConnections  *simplelru.LRU // Actual type is map[string]int64 which is "connId -> lastSeen"
	liveConnections        map[string]bool
	maxLiveStreams         int
	staleConnectionTimeout time.Duration
	stats                  AssemblerStats
}

// Context
// The assembler context
type context struct {
	CaptureInfo gopacket.CaptureInfo
	Origin      api.Capture
}

func (c *context) GetCaptureInfo() gopacket.CaptureInfo {
	return c.CaptureInfo
}

func NewTcpAssembler(outputItems chan *api.OutputChannelItem, streamsMap api.TcpStreamMap, opts *TapOpts) *tcpAssembler {
	var emitter api.Emitter = &api.Emitting{
		AppStats:      &diagnose.AppStats,
		OutputChannel: outputItems,
	}

	lastClosedConnections, _ := simplelru.NewLRU(lastClosedConnectionsMaxItems, func(key interface{}, value interface{}) {})

	a := &tcpAssembler{
		ignoredPorts:           opts.IgnoredPorts,
		lastClosedConnections:  lastClosedConnections,
		liveConnections:        make(map[string]bool),
		maxLiveStreams:         opts.maxLiveStreams,
		staleConnectionTimeout: opts.staleConnectionTimeout,
		stats:                  AssemblerStats{},
	}

	a.streamFactory = NewTcpStreamFactory(emitter, streamsMap, opts, a)
	a.streamPool = reassembly.NewStreamPool(a.streamFactory)
	a.Assembler = reassembly.NewAssembler(a.streamPool)

	maxBufferedPagesTotal := GetMaxBufferedPagesPerConnection()
	maxBufferedPagesPerConnection := GetMaxBufferedPagesTotal()
	logger.Log.Infof("Assembler options: maxBufferedPagesTotal=%d, maxBufferedPagesPerConnection=%d, opts=%v",
		maxBufferedPagesTotal, maxBufferedPagesPerConnection, opts)
	a.Assembler.AssemblerOptions.MaxBufferedPagesTotal = maxBufferedPagesTotal
	a.Assembler.AssemblerOptions.MaxBufferedPagesPerConnection = maxBufferedPagesPerConnection

	return a
}

func (a *tcpAssembler) processPackets(dumpPacket bool, packets <-chan source.TcpPacketInfo) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	ticker := time.NewTicker(a.staleConnectionTimeout)

out:
	for {
		select {
		case packetInfo := <-packets:
			if a.processPacket(packetInfo, dumpPacket) {
				break out
			}
		case <-signalChan:
			logger.Log.Infof("Caught SIGINT: aborting")
			break out
		case <-ticker.C:
			flushed, closed := a.FlushCloseOlderThan(time.Now().Add(-a.staleConnectionTimeout))
			stats := a.stats
			stats.closedConnections += closed
			stats.flushedConnections += flushed
		}
	}

	closed := a.FlushAll()
	logger.Log.Debugf("Final flush: %d closed", closed)
}

func (a *tcpAssembler) processPacket(packetInfo source.TcpPacketInfo, dumpPacket bool) bool {
	packetsCount := diagnose.AppStats.IncPacketsCount()

	if packetsCount%packetsSeenLogThreshold == 0 {
		logger.Log.Debugf("Packets seen: #%d", packetsCount)
	}

	packet := packetInfo.Packet
	data := packet.Data()
	diagnose.AppStats.UpdateProcessedBytes(uint64(len(data)))
	if dumpPacket {
		logger.Log.Debugf("Packet content (%d/0x%x) - %s", len(data), len(data), hex.Dump(data))
	}

	tcp := packet.Layer(layers.LayerTypeTCP)
	if tcp != nil {
		a.processTcpPacket(packetInfo.Source.Origin, packet, tcp.(*layers.TCP))
	}

	done := *maxcount > 0 && int64(diagnose.AppStats.PacketsCount) >= *maxcount
	if done {
		errorMapLen, _ := diagnose.TapErrors.GetErrorsSummary()
		logger.Log.Infof("Processed %v packets (%v bytes) in %v (errors: %v, errTypes:%v)",
			diagnose.AppStats.PacketsCount,
			diagnose.AppStats.ProcessedBytes,
			time.Since(diagnose.AppStats.StartTime),
			diagnose.TapErrors.ErrorsCount,
			errorMapLen)
	}
	return done
}

func (a *tcpAssembler) processTcpPacket(origin api.Capture, packet gopacket.Packet, tcp *layers.TCP) {
	diagnose.AppStats.IncTcpPacketsCount()
	if a.shouldIgnorePort(uint16(tcp.DstPort)) || a.shouldIgnorePort(uint16(tcp.SrcPort)) {
		diagnose.AppStats.IncIgnoredPacketsCount()
		return
	}

	id := getConnectionId(packet.NetworkLayer().NetworkFlow().Src().String(),
		packet.TransportLayer().TransportFlow().Src().String(),
		packet.NetworkLayer().NetworkFlow().Dst().String(),
		packet.TransportLayer().TransportFlow().Dst().String())

	if a.isLastAck(id) {
		diagnose.AppStats.IncIgnoredLastAckCount()
		return
	}

	if a.shouldThrottle(id) {
		diagnose.AppStats.IncThrottledPackets()
		return
	}

	c := context{
		CaptureInfo: packet.Metadata().CaptureInfo,
		Origin:      origin,
	}
	diagnose.InternalStats.Totalsz += len(tcp.Payload)
	if !dbgctl.MizuTapperDisableTcpReassembly {
		a.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, &c)
	}
}

func (a *tcpAssembler) tcpStreamCreated(stream *tcpStream) {
	a.liveConnections[stream.connectionId] = true
}

func (a *tcpAssembler) tcpStreamClosed(stream *tcpStream) {
	a.lastClosedConnections.Add(stream.connectionId, time.Now().UnixMilli())
	delete(a.liveConnections, stream.connectionId)
}

func (a *tcpAssembler) isLastAck(connectionId string) bool {
	if closedTimeMillis, ok := a.lastClosedConnections.Get(connectionId); ok {
		timeSinceClosed := time.Since(time.UnixMilli(closedTimeMillis.(int64)))
		if timeSinceClosed < lastAckThreshold {
			return true
		}
	}
	return false
}

func (a *tcpAssembler) shouldThrottle(connectionId string) bool {
	if _, ok := a.liveConnections[connectionId]; ok {
		return false
	}

	return len(a.liveConnections) > a.maxLiveStreams
}

func (a *tcpAssembler) dumpStreamPool() {
	a.streamPool.Dump()
}

func (a *tcpAssembler) waitAndDump() {
	a.streamFactory.WaitGoRoutines()
	logger.Log.Debugf("%s", a.Dump())
}

func (a *tcpAssembler) shouldIgnorePort(port uint16) bool {
	for _, p := range a.ignoredPorts {
		if port == p {
			return true
		}
	}

	return false
}

func (a *tcpAssembler) DumpStats() AssemblerStats {
	result := a.stats
	a.stats = AssemblerStats{}
	return result
}

func getConnectionId(saddr string, sport string, daddr string, dport string) string {
	s := fmt.Sprintf("%s:%s", saddr, sport)
	d := fmt.Sprintf("%s:%s", daddr, dport)
	if s > d {
		return fmt.Sprintf("%s#%s", s, d)
	} else {
		return fmt.Sprintf("%s#%s", d, s)
	}
}
