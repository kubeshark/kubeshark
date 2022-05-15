package tap

import (
	"encoding/binary"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/google/gopacket/reassembly"
	"github.com/up9inc/mizu/tap/diagnose"
)

type tcpReassemblyStream struct {
	ident      string
	tcpState   *reassembly.TCPSimpleFSM
	fsmerr     bool
	optchecker reassembly.TCPOptionCheck
	isDNS      bool
	tcpStream  *tcpStream
}

func NewTcpReassemblyStream(ident string, tcp *layers.TCP, fsmOptions reassembly.TCPSimpleFSMOptions, stream *tcpStream) reassembly.Stream {
	return &tcpReassemblyStream{
		ident:      ident,
		tcpState:   reassembly.NewTCPSimpleFSM(fsmOptions),
		optchecker: reassembly.NewTCPOptionCheck(),
		isDNS:      tcp.SrcPort == 53 || tcp.DstPort == 53,
		tcpStream:  stream,
	}
}

func (t *tcpReassemblyStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	// FSM
	if !t.tcpState.CheckState(tcp, dir) {
		diagnose.TapErrors.SilentError("FSM-rejection", "%s: Packet rejected by FSM (state:%s)", t.ident, t.tcpState.String())
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

	*start = true

	return accept
}

func (t *tcpReassemblyStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
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

	if skip != -1 && skip != 0 {
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
	} else if t.tcpStream.GetIsTapTarget() {
		if length > 0 {
			// This is where we pass the reassembled information onwards
			// This channel is read by an tcpReader object
			diagnose.AppStats.IncReassembledTcpPayloadsCount()
			timestamp := ac.GetCaptureInfo().Timestamp
			if dir == reassembly.TCPDirClientToServer {
				t.tcpStream.client.sendMsgIfNotClosed(NewTcpReaderDataMsg(data, timestamp))
			} else {
				t.tcpStream.server.sendMsgIfNotClosed(NewTcpReaderDataMsg(data, timestamp))
			}
		}
	}
}

func (t *tcpReassemblyStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	if t.tcpStream.GetIsTapTarget() && !t.tcpStream.GetIsClosed() {
		t.tcpStream.close()
	}
	// do not remove the connection to allow last ACK
	return false
}
