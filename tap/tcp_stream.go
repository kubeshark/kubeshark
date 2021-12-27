package tap

import (
	"encoding/binary"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/google/gopacket/reassembly"
	"github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/diagnose"
)

/* It's a connection (bidirectional)
 * Implements gopacket.reassembly.Stream interface (Accept, ReassembledSG, ReassemblyComplete)
 * ReassembledSG gets called when new reassembled data is ready (i.e. bytes in order, no duplicates, complete)
 * In our implementation, we pass information from ReassembledSG to the tcpReader through a shared channel.
 */
type tcpStream struct {
	id              int64
	isClosed        bool
	superIdentifier *api.SuperIdentifier
	tcpstate        *reassembly.TCPSimpleFSM
	fsmerr          bool
	optchecker      reassembly.TCPOptionCheck
	net, transport  gopacket.Flow
	isDNS           bool
	isTapTarget     bool
	clients         []tcpReader
	servers         []tcpReader
	ident           string
	sync.Mutex
	streamsMap *tcpStreamMap
}

func (t *tcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	// FSM
	if !t.tcpstate.CheckState(tcp, dir) {
		diagnose.TapErrors.SilentError("FSM-rejection", "%s: Packet rejected by FSM (state:%s)", t.ident, t.tcpstate.String())
		diagnose.InternalStats.RejectFsm++
		if !t.fsmerr {
			t.fsmerr = true
			diagnose.InternalStats.RejectConnFsm++
		}
		if !*ignorefsmerr {
			return false
		}
	}
	// Options
	err := t.optchecker.Accept(tcp, ci, dir, nextSeq, start)
	if err != nil {
		diagnose.TapErrors.SilentError("OptionChecker-rejection", "%s: Packet rejected by OptionChecker: %s", t.ident, err)
		diagnose.InternalStats.RejectOpt++
		if !*nooptcheck {
			return false
		}
	}
	// Checksum
	accept := true
	if *checksum {
		c, err := tcp.ComputeChecksum()
		if err != nil {
			diagnose.TapErrors.SilentError("ChecksumCompute", "%s: Got error computing checksum: %s", t.ident, err)
			accept = false
		} else if c != 0x0 {
			diagnose.TapErrors.SilentError("Checksum", "%s: Invalid checksum: 0x%x", t.ident, c)
			accept = false
		}
	}
	if !accept {
		diagnose.InternalStats.RejectOpt++
	}
	return accept
}

func (t *tcpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	dir, _, _, skip := sg.Info()
	length, saved := sg.Lengths()
	// update stats
	sgStats := sg.Stats()
	if skip > 0 {
		diagnose.InternalStats.MissedBytes += skip
	}
	diagnose.InternalStats.Sz += length - saved
	diagnose.InternalStats.Pkt += sgStats.Packets
	if sgStats.Chunks > 1 {
		diagnose.InternalStats.Reassembled++
	}
	diagnose.InternalStats.OutOfOrderPackets += sgStats.QueuedPackets
	diagnose.InternalStats.OutOfOrderBytes += sgStats.QueuedBytes
	if length > diagnose.InternalStats.BiggestChunkBytes {
		diagnose.InternalStats.BiggestChunkBytes = length
	}
	if sgStats.Packets > diagnose.InternalStats.BiggestChunkPackets {
		diagnose.InternalStats.BiggestChunkPackets = sgStats.Packets
	}
	if sgStats.OverlapBytes != 0 && sgStats.OverlapPackets == 0 {
		// In the original example this was handled with panic().
		// I don't know what this error means or how to handle it properly.
		diagnose.TapErrors.SilentError("Invalid-Overlap", "bytes:%d, pkts:%d", sgStats.OverlapBytes, sgStats.OverlapPackets)
	}
	diagnose.InternalStats.OverlapBytes += sgStats.OverlapBytes
	diagnose.InternalStats.OverlapPackets += sgStats.OverlapPackets

	if skip == -1 && *allowmissinginit {
		// this is allowed
	} else if skip != 0 {
		// Missing bytes in stream: do not even try to parse it
		return
	}
	data := sg.Fetch(length)
	if t.isDNS {
		dns := &layers.DNS{}
		var decoded []gopacket.LayerType
		if len(data) < 2 {
			if len(data) > 0 {
				sg.KeepFrom(0)
			}
			return
		}
		dnsSize := binary.BigEndian.Uint16(data[:2])
		missing := int(dnsSize) - len(data[2:])
		diagnose.TapErrors.Debug("dnsSize: %d, missing: %d", dnsSize, missing)
		if missing > 0 {
			diagnose.TapErrors.Debug("Missing some bytes: %d", missing)
			sg.KeepFrom(0)
			return
		}
		p := gopacket.NewDecodingLayerParser(layers.LayerTypeDNS, dns)
		err := p.DecodeLayers(data[2:], &decoded)
		if err != nil {
			diagnose.TapErrors.SilentError("DNS-parser", "Failed to decode DNS: %v", err)
		} else {
			diagnose.TapErrors.Debug("DNS: %s", gopacket.LayerDump(dns))
		}
		if len(data) > 2+int(dnsSize) {
			sg.KeepFrom(2 + int(dnsSize))
		}
	} else if t.isTapTarget {
		if length > 0 {
			// This is where we pass the reassembled information onwards
			// This channel is read by an tcpReader object
			diagnose.AppStats.IncReassembledTcpPayloadsCount()
			timestamp := ac.GetCaptureInfo().Timestamp
			if dir == reassembly.TCPDirClientToServer {
				for i := range t.clients {
					reader := &t.clients[i]
					reader.Lock()
					if !reader.isClosed {
						reader.msgQueue <- tcpReaderDataMsg{data, timestamp}
					}
					reader.Unlock()
				}
			} else {
				for i := range t.servers {
					reader := &t.servers[i]
					reader.Lock()
					if !reader.isClosed {
						reader.msgQueue <- tcpReaderDataMsg{data, timestamp}
					}
					reader.Unlock()
				}
			}
		}
	}
}

func (t *tcpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	if t.isTapTarget && !t.isClosed {
		t.Close()
	}
	// do not remove the connection to allow last ACK
	return false
}

func (t *tcpStream) Close() {
	shouldReturn := false
	t.Lock()
	if t.isClosed {
		shouldReturn = true
	} else {
		t.isClosed = true
	}
	t.Unlock()
	if shouldReturn {
		return
	}
	t.streamsMap.Delete(t.id)

	for i := range t.clients {
		reader := &t.clients[i]
		reader.Close()
	}
	for i := range t.servers {
		reader := &t.servers[i]
		reader.Close()
	}
}
