package tap

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/google/gopacket/reassembly"
	"github.com/up9inc/mizu/tap/api"
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
		tapErrors.SilentError("FSM-rejection", "%s: Packet rejected by FSM (state:%s)", t.ident, t.tcpstate.String())
		internalStats.rejectFsm++
		if !t.fsmerr {
			t.fsmerr = true
			internalStats.rejectConnFsm++
		}
		if !*ignorefsmerr {
			return false
		}
	}
	// Options
	err := t.optchecker.Accept(tcp, ci, dir, nextSeq, start)
	if err != nil {
		tapErrors.SilentError("OptionChecker-rejection", "%s: Packet rejected by OptionChecker: %s", t.ident, err)
		internalStats.rejectOpt++
		if !*nooptcheck {
			return false
		}
	}
	// Checksum
	accept := true
	if *checksum {
		c, err := tcp.ComputeChecksum()
		if err != nil {
			tapErrors.SilentError("ChecksumCompute", "%s: Got error computing checksum: %s", t.ident, err)
			accept = false
		} else if c != 0x0 {
			tapErrors.SilentError("Checksum", "%s: Invalid checksum: 0x%x", t.ident, c)
			accept = false
		}
	}
	if !accept {
		internalStats.rejectOpt++
	}
	return accept
}

func (t *tcpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	dir, start, end, skip := sg.Info()
	length, saved := sg.Lengths()
	// update stats
	sgStats := sg.Stats()
	if skip > 0 {
		internalStats.missedBytes += skip
	}
	internalStats.sz += length - saved
	internalStats.pkt += sgStats.Packets
	if sgStats.Chunks > 1 {
		internalStats.reassembled++
	}
	internalStats.outOfOrderPackets += sgStats.QueuedPackets
	internalStats.outOfOrderBytes += sgStats.QueuedBytes
	if length > internalStats.biggestChunkBytes {
		internalStats.biggestChunkBytes = length
	}
	if sgStats.Packets > internalStats.biggestChunkPackets {
		internalStats.biggestChunkPackets = sgStats.Packets
	}
	if sgStats.OverlapBytes != 0 && sgStats.OverlapPackets == 0 {
		// In the original example this was handled with panic().
		// I don't know what this error means or how to handle it properly.
		tapErrors.SilentError("Invalid-Overlap", "bytes:%d, pkts:%d", sgStats.OverlapBytes, sgStats.OverlapPackets)
	}
	internalStats.overlapBytes += sgStats.OverlapBytes
	internalStats.overlapPackets += sgStats.OverlapPackets

	var ident string
	if dir == reassembly.TCPDirClientToServer {
		ident = fmt.Sprintf("%v %v(%s): ", t.net, t.transport, dir)
	} else {
		ident = fmt.Sprintf("%v %v(%s): ", t.net.Reverse(), t.transport.Reverse(), dir)
	}
	tapErrors.Debug("%s: SG reassembled packet with %d bytes (start:%v,end:%v,skip:%d,saved:%d,nb:%d,%d,overlap:%d,%d)", ident, length, start, end, skip, saved, sgStats.Packets, sgStats.Chunks, sgStats.OverlapBytes, sgStats.OverlapPackets)
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
		tapErrors.Debug("dnsSize: %d, missing: %d", dnsSize, missing)
		if missing > 0 {
			tapErrors.Debug("Missing some bytes: %d", missing)
			sg.KeepFrom(0)
			return
		}
		p := gopacket.NewDecodingLayerParser(layers.LayerTypeDNS, dns)
		err := p.DecodeLayers(data[2:], &decoded)
		if err != nil {
			tapErrors.SilentError("DNS-parser", "Failed to decode DNS: %v", err)
		} else {
			tapErrors.Debug("DNS: %s", gopacket.LayerDump(dns))
		}
		if len(data) > 2+int(dnsSize) {
			sg.KeepFrom(2 + int(dnsSize))
		}
	} else if t.isTapTarget {
		if length > 0 {
			// This is where we pass the reassembled information onwards
			// This channel is read by an tcpReader object
			appStats.IncReassembledTcpPayloadsCount()
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
	tapErrors.Debug("%s: Connection closed", t.ident)
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
