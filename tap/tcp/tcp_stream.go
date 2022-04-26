package tcp

import (
	"encoding/binary"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/google/gopacket/reassembly"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/diagnose"
)

var checksum = flag.Bool("checksum", false, "Check TCP checksum")                                                      // global
var nooptcheck = flag.Bool("nooptcheck", true, "Do not check TCP options (useful to ignore MSS on captures with TSO)") // global
var ignorefsmerr = flag.Bool("ignorefsmerr", true, "Ignore TCP FSM errors")                                            // global

/* It's a connection (bidirectional)
 * Implements gopacket.reassembly.Stream interface (Accept, ReassembledSG, ReassemblyComplete)
 * ReassembledSG gets called when new reassembled data is ready (i.e. bytes in order, no duplicates, complete)
 * In our implementation, we pass information from ReassembledSG to the TcpReader through a shared channel.
 */
type tcpStream struct {
	id              int64
	isClosed        bool
	protoIdentifier *api.ProtoIdentifier
	tcpState        *reassembly.TCPSimpleFSM
	fsmerr          bool
	optchecker      reassembly.TCPOptionCheck
	net, transport  gopacket.Flow
	isDNS           bool
	isTapTarget     bool
	clients         []api.TcpReader
	servers         []api.TcpReader
	ident           string
	origin          api.Capture
	reqResMatcher   api.RequestResponseMatcher
	createdAt       time.Time
	streamsMap      api.TcpStreamMap
	sync.Mutex
}

func NewTcpStream(net gopacket.Flow, transport gopacket.Flow, tcp *layers.TCP, isTapTarget bool, fsmOptions reassembly.TCPSimpleFSMOptions, streamsMap api.TcpStreamMap, capture api.Capture) api.TcpStream {
	return &tcpStream{
		net:             net,
		transport:       transport,
		isDNS:           tcp.SrcPort == 53 || tcp.DstPort == 53,
		isTapTarget:     isTapTarget,
		tcpState:        reassembly.NewTCPSimpleFSM(fsmOptions),
		ident:           fmt.Sprintf("%s:%s", net, transport),
		optchecker:      reassembly.NewTCPOptionCheck(),
		protoIdentifier: &api.ProtoIdentifier{},
		streamsMap:      streamsMap,
		origin:          capture,
	}
}

func NewTcpStreamDummy(capture api.Capture) api.TcpStream {
	return &tcpStream{
		origin:          capture,
		protoIdentifier: &api.ProtoIdentifier{},
	}
}

func (t *tcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
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
	} else if t.isTapTarget {
		if length > 0 {
			// This is where we pass the reassembled information onwards
			// This channel is read by an tcpReader object
			diagnose.AppStatsInst.IncReassembledTcpPayloadsCount()
			timestamp := ac.GetCaptureInfo().Timestamp
			if dir == reassembly.TCPDirClientToServer {
				for i := range t.clients {
					reader := t.clients[i]
					reader.SendMsgIfNotClosed(NewTcpReaderDataMsg(data, timestamp))
				}
			} else {
				for i := range t.servers {
					reader := t.servers[i]
					reader.SendMsgIfNotClosed(NewTcpReaderDataMsg(data, timestamp))
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
	t.Lock()
	defer t.Unlock()

	if t.isClosed {
		return
	}

	t.isClosed = true

	t.streamsMap.Delete(t.id)

	for i := range t.clients {
		reader := t.clients[i]
		reader.Close()
	}
	for i := range t.servers {
		reader := t.servers[i]
		reader.Close()
	}
}

func (t *tcpStream) CloseOtherProtocolDissectors(protocol *api.Protocol) {
	t.Lock()
	defer t.Unlock()

	if t.protoIdentifier.IsClosedOthers {
		return
	}

	t.protoIdentifier.Protocol = protocol

	for i := range t.clients {
		reader := t.clients[i]
		if reader.GetExtension().Protocol != t.protoIdentifier.Protocol {
			reader.Close()
		}
	}
	for i := range t.servers {
		reader := t.servers[i]
		if reader.GetExtension().Protocol != t.protoIdentifier.Protocol {
			reader.Close()
		}
	}

	t.protoIdentifier.IsClosedOthers = true
}

func (t *tcpStream) AddClient(reader api.TcpReader) {
	t.clients = append(t.clients, reader)
}

func (t *tcpStream) AddServer(reader api.TcpReader) {
	t.servers = append(t.servers, reader)
}

func (t *tcpStream) ClientRun(index int, filteringOptions *shared.TrafficFilteringOptions, wg *sync.WaitGroup) {
	t.clients[index].Run(filteringOptions, wg)
}

func (t *tcpStream) ServerRun(index int, filteringOptions *shared.TrafficFilteringOptions, wg *sync.WaitGroup) {
	t.servers[index].Run(filteringOptions, wg)
}

func (t *tcpStream) GetOrigin() api.Capture {
	return t.origin
}

func (t *tcpStream) GetProtoIdentifier() *api.ProtoIdentifier {
	return t.protoIdentifier
}

func (t *tcpStream) GetReqResMatcher() api.RequestResponseMatcher {
	return t.reqResMatcher
}

func (t *tcpStream) GetIsTapTarget() bool {
	return t.isTapTarget
}

func (t *tcpStream) GetId() int64 {
	return t.id
}

func (t *tcpStream) SetId(id int64) {
	t.id = id
}
