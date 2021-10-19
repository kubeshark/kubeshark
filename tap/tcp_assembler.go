package tap

import (
	"encoding/hex"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
)

type tcpAssembler struct {
	*reassembly.Assembler
	streamPool     *reassembly.StreamPool
	streamFactory  *tcpStreamFactory
	assemblerMutex sync.Mutex
}

func NewTcpAssember(outputItems chan *api.OutputChannelItem, streamsMap *tcpStreamMap) *tcpAssembler {
	var emitter api.Emitter = &api.Emitting{
		AppStats:      &appStats,
		OutputChannel: outputItems,
	}

	streamFactory := NewTcpStreamFactory(emitter, streamsMap)
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

func (a *tcpAssembler) processPackets(dumpPacket bool, packets <-chan tcpPacketInfo) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	for packetInfo := range packets {
		packetsCount := appStats.IncPacketsCount()
		logger.Log.Debugf("PACKET #%d", packetsCount)
		packet := packetInfo.packet
		data := packet.Data()
		appStats.UpdateProcessedBytes(uint64(len(data)))
		if dumpPacket {
			logger.Log.Debugf("Packet content (%d/0x%x) - %s", len(data), len(data), hex.Dump(data))
		}

		tcp := packet.Layer(layers.LayerTypeTCP)
		if tcp != nil {
			appStats.IncTcpPacketsCount()
			tcp := tcp.(*layers.TCP)
			if *checksum {
				err := tcp.SetNetworkLayerForChecksum(packet.NetworkLayer())
				if err != nil {
					logger.Log.Fatalf("Failed to set network layer for checksum: %s\n", err)
				}
			}
			c := Context{
				CaptureInfo: packet.Metadata().CaptureInfo,
			}
			stats.totalsz += len(tcp.Payload)
			logger.Log.Debugf("%s : %v -> %s : %v", packet.NetworkLayer().NetworkFlow().Src(), tcp.SrcPort, packet.NetworkLayer().NetworkFlow().Dst(), tcp.DstPort)
			a.assemblerMutex.Lock()
			a.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, &c)
			a.assemblerMutex.Unlock()
		}

		done := *maxcount > 0 && int64(appStats.PacketsCount) >= *maxcount
		if done {
			errorMapLen, _ := tapErrors.getErrorsSummary()
			logger.Log.Infof("Processed %v packets (%v bytes) in %v (errors: %v, errTypes:%v)",
				appStats.PacketsCount,
				appStats.ProcessedBytes,
				time.Since(appStats.StartTime),
				tapErrors.nErrors,
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
