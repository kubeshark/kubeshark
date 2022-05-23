package tap

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"github.com/struCoder/pidusage"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/dbgctl"
	"github.com/up9inc/mizu/tap/diagnose"
	"github.com/up9inc/mizu/tap/source"
)

const PACKETS_SEEN_LOG_THRESHOLD = 1000

type tcpAssembler struct {
	*reassembly.Assembler
	streamPool      *reassembly.StreamPool
	streamFactory   *tcpStreamFactory
	assemblerMutex  sync.Mutex
	ignoredPorts    []uint16
	liveStreams     map[string]bool
	liveStreamsLock sync.RWMutex
	sysInfo         *pidusage.SysInfo
	cpuLimit        float64
	tapperPid       int
}

// Context
// The assembler context
type context struct {
	CaptureInfo  gopacket.CaptureInfo
	Origin       api.Capture
	connectionId string
}

func (c *context) GetCaptureInfo() gopacket.CaptureInfo {
	return c.CaptureInfo
}

func NewTcpAssembler(outputItems chan *api.OutputChannelItem, streamsMap *tcpStreamMap, opts *TapOpts) *tcpAssembler {
	var emitter api.Emitter = &api.Emitting{
		AppStats:      &diagnose.AppStats,
		OutputChannel: outputItems,
	}

	a := &tcpAssembler{
		ignoredPorts: opts.IgnoredPorts,
		cpuLimit:     opts.cpuLimit,
		liveStreams:  make(map[string]bool),
		tapperPid:    os.Getpid(),
		sysInfo:      &pidusage.SysInfo{CPU: -1, Memory: -1},
	}

	closeHandler := func(stream *tcpStream) {
		a.liveStreamsLock.Lock()
		defer a.liveStreamsLock.Unlock()
		delete(a.liveStreams, stream.connectionId)
	}

	createdHandler := func(stream *tcpStream) {
		a.liveStreamsLock.Lock()
		defer a.liveStreamsLock.Unlock()
		a.liveStreams[stream.connectionId] = true
	}

	streamFactory := NewTcpStreamFactory(emitter, streamsMap, opts, closeHandler, createdHandler)

	streamPool := reassembly.NewStreamPool(streamFactory)
	assembler := reassembly.NewAssembler(streamPool)

	maxBufferedPagesTotal := GetMaxBufferedPagesPerConnection()
	maxBufferedPagesPerConnection := GetMaxBufferedPagesTotal()
	logger.Log.Infof("Assembler options: maxBufferedPagesTotal=%d, maxBufferedPagesPerConnection=%d, opts=%v",
		maxBufferedPagesTotal, maxBufferedPagesPerConnection, opts)
	assembler.AssemblerOptions.MaxBufferedPagesTotal = maxBufferedPagesTotal
	assembler.AssemblerOptions.MaxBufferedPagesPerConnection = maxBufferedPagesPerConnection

	a.streamPool = streamPool
	a.streamFactory = streamFactory
	a.Assembler = assembler

	return a
}

func (a *tcpAssembler) buildConnectionId(saddr string, daddr string, sport string, dport string) string {
	s := fmt.Sprintf("%s:%s", saddr, sport)
	d := fmt.Sprintf("%s:%s", daddr, dport)
	if s > d {
		return fmt.Sprintf("%s#%s", s, d)
	} else {
		return fmt.Sprintf("%s#%s", d, s)
	}
}

func (a *tcpAssembler) shouldThrottleNewStreams(connectionId string) bool {
	if a.cpuLimit == 0 {
		return false
	}

	return a.sysInfo.CPU > a.cpuLimit
}

func (a *tcpAssembler) connectionExists(connectionId string) bool {
	a.liveStreamsLock.RLock()
	defer a.liveStreamsLock.RUnlock()
	_, ok := a.liveStreams[connectionId]
	return ok
}

func (a *tcpAssembler) handlePacket(packetInfo *source.TcpPacketInfo, dumpPacket bool) bool {
	packetsCount := diagnose.AppStats.IncPacketsCount()

	if packetsCount%PACKETS_SEEN_LOG_THRESHOLD == 0 {
		logger.Log.Debugf("Packets seen: #%d", packetsCount)
	}

	packet := packetInfo.Packet
	data := packet.Data()
	diagnose.AppStats.UpdateProcessedBytes(uint64(len(data)))
	if dumpPacket {
		logger.Log.Debugf("Packet content (%d/0x%x) - %s", len(data), len(data), hex.Dump(data))
	}

	done := *maxcount > 0 && int64(diagnose.AppStats.PacketsCount) >= *maxcount

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return done
	}

	diagnose.AppStats.IncTcpPacketsCount()

	tcp := tcpLayer.(*layers.TCP)
	if a.shouldIgnorePort(uint16(tcp.DstPort)) || a.shouldIgnorePort(uint16(tcp.SrcPort)) {
		diagnose.AppStats.IncIgnoredPacketsCount()
		return done
	}
	if dbgctl.MizuTapperDisableTcpReassembly {
		return done
	}

	connectionId := a.buildConnectionId(packet.NetworkLayer().NetworkFlow().Src().String(),
		packet.NetworkLayer().NetworkFlow().Dst().String(),
		packet.TransportLayer().TransportFlow().Src().String(),
		packet.TransportLayer().TransportFlow().Dst().String())

	if !a.connectionExists(connectionId) {
		if a.shouldThrottleNewStreams(connectionId) {
			diagnose.AppStats.IncThrottledPackets()
			return done
		}
	}

	c := context{
		CaptureInfo:  packet.Metadata().CaptureInfo,
		Origin:       packetInfo.Source.Origin,
		connectionId: connectionId,
	}
	diagnose.InternalStats.Totalsz += len(tcp.Payload)
	a.assemblerMutex.Lock()
	a.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, &c)
	a.assemblerMutex.Unlock()

	return done
}

func (a *tcpAssembler) processPackets(dumpPacket bool, packets <-chan source.TcpPacketInfo) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	ticker := time.NewTicker(1000 * time.Millisecond)

out:
	for {
		select {
		case <-ticker.C:
			a.updateUsage()
		case packetInfo := <-packets:
			if a.handlePacket(&packetInfo, dumpPacket) {
				break out
			}
		case <-signalChan:
			logger.Log.Infof("Caught SIGINT: aborting")
			break out
		}
	}

	errorMapLen, _ := diagnose.TapErrors.GetErrorsSummary()
	logger.Log.Infof("Processed %v packets (%v bytes) in %v (errors: %v, errTypes:%v)",
		diagnose.AppStats.PacketsCount,
		diagnose.AppStats.ProcessedBytes,
		time.Since(diagnose.AppStats.StartTime),
		diagnose.TapErrors.ErrorsCount,
		errorMapLen)

	a.assemblerMutex.Lock()
	closed := a.FlushAll()
	a.assemblerMutex.Unlock()
	logger.Log.Debugf("Final flush: %d closed", closed)
}

func (a *tcpAssembler) dumpStreamPool() {
	a.streamPool.Dump()
}

func (a *tcpAssembler) waitAndDump() {
	a.streamFactory.WaitGoRoutines()
	a.assemblerMutex.Lock()
	logger.Log.Debugf("%s", a.Dump())
	a.assemblerMutex.Unlock()
}

func (a *tcpAssembler) shouldIgnorePort(port uint16) bool {
	for _, p := range a.ignoredPorts {
		if port == p {
			return true
		}
	}

	return false
}

func (a *tcpAssembler) updateUsage() {
	sysInfo, err := pidusage.GetStat(a.tapperPid)

	if err != nil {
		logger.Log.Warningf("Unable to get CPU Usage for %d", a.tapperPid)
		a.sysInfo = &pidusage.SysInfo{
			CPU:    -1,
			Memory: -1,
		}
		return
	}

	a.sysInfo = sysInfo
}
