package api

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/google/gopacket/reassembly"
	"github.com/up9inc/mizu/tap/api/diagnose"
)

/* It's a connection (bidirectional)
 * Implements gopacket.reassembly.Stream interface (Accept, ReassembledSG, ReassemblyComplete)
 * ReassembledSG gets called when new reassembled data is ready (i.e. bytes in order, no duplicates, complete)
 * In our implementation, we pass information from ReassembledSG to the TcpReader through a shared channel.
 */
type TcpStream struct {
	Id              int64
	isClosed        bool
	ProtoIdentifier *ProtoIdentifier
	TcpState        *reassembly.TCPSimpleFSM
	fsmerr          bool
	Optchecker      reassembly.TCPOptionCheck
	Net, Transport  gopacket.Flow
	IsDNS           bool
	IsTapTarget     bool
	Clients         []TcpReader
	Servers         []TcpReader
	Ident           string
	Origin          Capture
	ReqResMatcher   RequestResponseMatcher
	createdAt       time.Time
	StreamsMap      *TcpStreamMap
	sync.Mutex
}

func (t *TcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	// FSM
	if !t.TcpState.CheckState(tcp, dir) {
		diagnose.TapErrors.SilentError("FSM-rejection", "%s: Packet rejected by FSM (state:%s)", t.Ident, t.TcpState.String())
		diagnose.InternalStats.RejectFsm++
		if !t.fsmerr {
			t.fsmerr = true
			diagnose.InternalStats.RejectConnFsm++
		}
	}
	// Options
	err := t.Optchecker.Accept(tcp, ci, dir, nextSeq, start)
	if err != nil {
		diagnose.TapErrors.SilentError("OptionChecker-rejection", "%s: Packet rejected by OptionChecker: %s", t.Ident, err)
		diagnose.InternalStats.RejectOpt++
	}

	*start = true

	return true
}

func (t *TcpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
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
	if t.IsDNS {
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
	} else if t.IsTapTarget {
		if length > 0 {
			// This is where we pass the reassembled information onwards
			// This channel is read by an tcpReader object
			diagnose.AppStatsInst.IncReassembledTcpPayloadsCount()
			timestamp := ac.GetCaptureInfo().Timestamp
			if dir == reassembly.TCPDirClientToServer {
				for i := range t.Clients {
					reader := &t.Clients[i]
					reader.Lock()
					if !reader.isClosed {
						reader.MsgQueue <- TcpReaderDataMsg{data, timestamp}
					}
					reader.Unlock()
				}
			} else {
				for i := range t.Servers {
					reader := &t.Servers[i]
					reader.Lock()
					if !reader.isClosed {
						reader.MsgQueue <- TcpReaderDataMsg{data, timestamp}
					}
					reader.Unlock()
				}
			}
		}
	}
}

func (t *TcpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	if t.IsTapTarget && !t.isClosed {
		t.Close()
	}
	// do not remove the connection to allow last ACK
	return false
}

func (t *TcpStream) Close() {
	t.Lock()
	defer t.Unlock()

	if t.isClosed {
		return
	}

	t.isClosed = true

	t.StreamsMap.Delete(t.Id)

	for i := range t.Clients {
		reader := &t.Clients[i]
		reader.Close()
	}
	for i := range t.Servers {
		reader := &t.Servers[i]
		reader.Close()
	}
}

func (t *TcpStream) CloseOtherProtocolDissectors(protocol *Protocol) {
	t.Lock()
	defer t.Unlock()

	if t.ProtoIdentifier.IsClosedOthers {
		return
	}

	t.ProtoIdentifier.Protocol = protocol

	for i := range t.Clients {
		reader := &t.Clients[i]
		if reader.Extension.Protocol != t.ProtoIdentifier.Protocol {
			reader.Close()
		}
	}
	for i := range t.Servers {
		reader := &t.Servers[i]
		if reader.Extension.Protocol != t.ProtoIdentifier.Protocol {
			reader.Close()
		}
	}

	t.ProtoIdentifier.IsClosedOthers = true
}
