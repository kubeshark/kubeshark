package tap

import (
	"encoding/hex"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/diagnose"
	"github.com/up9inc/mizu/tap/source"
)

const PACKETS_SEEN_LOG_THRESHOLD = 1000

type tcpAssembler struct {
	*reassembly.Assembler
	streamPool     *reassembly.StreamPool
	streamFactory  *tcpStreamFactory
	assemblerMutex sync.Mutex
}

// Context
// The assembler context
type context struct {
	CaptureInfo gopacket.CaptureInfo
}

func (c *context) GetCaptureInfo() gopacket.CaptureInfo {
	return c.CaptureInfo
}

func NewTcpAssembler(outputItems chan *api.OutputChannelItem, streamsMap *tcpStreamMap, opts *TapOpts) *tcpAssembler {
	var emitter api.Emitter = &api.Emitting{
		AppStats:      &diagnose.AppStats,
		OutputChannel: outputItems,
	}

	streamFactory := NewTcpStreamFactory(emitter, streamsMap, opts)
	streamPool := reassembly.NewStreamPool(streamFactory)
	assembler := reassembly.NewAssembler(streamPool)

	maxBufferedPagesTotal := GetMaxBufferedPagesPerConnection()
	maxBufferedPagesPerConnection := GetMaxBufferedPagesTotal()
	logger.Log.Infof("Assembler options: maxBufferedPagesTotal=%d, maxBufferedPagesPerConnection=%d",
		maxBufferedPagesTotal, maxBufferedPagesPerConnection)
	assembler.AssemblerOptions.MaxBufferedPagesTotal = maxBufferedPagesTotal
	assembler.AssemblerOptions.MaxBufferedPagesPerConnection = maxBufferedPagesPerConnection

	return &tcpAssembler{
		Assembler:     assembler,
		streamPool:    streamPool,
		streamFactory: streamFactory,
	}
}

func (a *tcpAssembler) processPackets(dumpPacket bool, packets <-chan source.TcpPacketInfo) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	for packetInfo := range packets {
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

		tcp := packet.Layer(layers.LayerTypeTCP)
		if tcp != nil {
			diagnose.AppStats.IncTcpPacketsCount()
			tcp := tcp.(*layers.TCP)
			if *checksum {
				err := tcp.SetNetworkLayerForChecksum(packet.NetworkLayer())
				if err != nil {
					logger.Log.Fatalf("Failed to set network layer for checksum: %s", err)
				}
			}
			c := context{
				CaptureInfo: packet.Metadata().CaptureInfo,
			}
			diagnose.InternalStats.Totalsz += len(tcp.Payload)
			a.assemblerMutex.Lock()
			a.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, &c)
			a.assemblerMutex.Unlock()
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

		select {
		case <-signalChan:
			logger.Log.Infof("Caught SIGINT: aborting")
			done = true
		default:
			// NOP: continue
		}
		if done {
			break
		}
	}

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
