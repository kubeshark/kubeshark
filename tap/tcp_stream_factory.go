package tap

import (
	"fmt"
	"sync"

	"github.com/romana/rlog"
	"github.com/up9inc/mizu/tap/api"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/google/gopacket/reassembly"
)

/*
 * The TCP factory: returns a new Stream
 * Implements gopacket.reassembly.StreamFactory interface (New)
 * Generates a new tcp stream for each new tcp connection. Closes the stream when the connection closes.
 */
type tcpStreamFactory struct {
	wg                 sync.WaitGroup
	outboundLinkWriter *OutboundLinkWriter
	AllExtensionPorts  []string
	Emitter            api.Emitter
}

func (factory *tcpStreamFactory) New(net, transport gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	rlog.Debugf("* NEW: %s %s", net, transport)
	fsmOptions := reassembly.TCPSimpleFSMOptions{
		SupportMissingEstablishment: *allowmissinginit,
	}
	rlog.Debugf("Current App Ports: %v", factory.AllExtensionPorts)
	srcIp := net.Src().String()
	dstIp := net.Dst().String()
	srcPort := transport.Src().String()
	dstPort := transport.Dst().String()

	// if factory.shouldNotifyOnOutboundLink(dstIp, dstPort) {
	// 	factory.outboundLinkWriter.WriteOutboundLink(net.Src().String(), dstIp, dstPort, "", "")
	// }
	props := factory.getStreamProps(srcIp, dstIp, dstPort)
	isTapTarget := props.isTapTarget
	stream := &tcpStream{
		net:         net,
		transport:   transport,
		isDNS:       tcp.SrcPort == 53 || tcp.DstPort == 53,
		isTapTarget: isTapTarget,
		tcpstate:    reassembly.NewTCPSimpleFSM(fsmOptions),
		ident:       fmt.Sprintf("%s:%s", net, transport),
		optchecker:  reassembly.NewTCPOptionCheck(),
	}
	if stream.isTapTarget {
		stream.reader = tcpReader{
			msgQueue: make(chan tcpReaderDataMsg),
			ident:    fmt.Sprintf("%s %s", net, transport),
			tcpID: &api.TcpID{
				SrcIP:   srcIp,
				DstIP:   dstIp,
				SrcPort: srcPort,
				DstPort: dstPort,
			},
			parent:             stream,
			isClient:           true,
			isOutgoing:         props.isOutgoing,
			outboundLinkWriter: factory.outboundLinkWriter,
			Emitter:            factory.Emitter,
		}
		factory.wg.Add(1)
		// Start reading from channel stream.reader.bytes
		go stream.reader.run(&factory.wg)
	}
	return stream
}

func (factory *tcpStreamFactory) WaitGoRoutines() {
	factory.wg.Wait()
}

func (factory *tcpStreamFactory) getStreamProps(srcIP string, dstIP string, dstPort string) *streamProps {
	if hostMode {
		if inArrayString(gSettings.filterAuthorities, fmt.Sprintf("%s:%s", dstIP, dstPort)) {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ host1 %s:%s", dstIP, dstPort))
			return &streamProps{isTapTarget: true, isOutgoing: false}
		} else if inArrayString(gSettings.filterAuthorities, dstIP) {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ host2 %s", dstIP))
			return &streamProps{isTapTarget: true, isOutgoing: false}
		} else if *anydirection && inArrayString(gSettings.filterAuthorities, srcIP) {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ host3 %s", srcIP))
			return &streamProps{isTapTarget: true, isOutgoing: true}
		}
		return &streamProps{isTapTarget: false}
	} else {
		isOutgoing := !inArrayString(ownIps, dstIP)

		if !*anydirection && isOutgoing {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("- notHost2"))
			return &streamProps{isTapTarget: false, isOutgoing: isOutgoing}
		}

		rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ notHost3 %s -> %s:%s", srcIP, dstIP, dstPort))
		return &streamProps{isTapTarget: true}
	}
}

func (factory *tcpStreamFactory) shouldNotifyOnOutboundLink(dstIP string, dstPort int) bool {
	if inArrayInt(remoteOnlyOutboundPorts, dstPort) {
		isDirectedHere := inArrayString(ownIps, dstIP)
		return !isDirectedHere && !isPrivateIP(dstIP)
	}
	return true
}

type streamProps struct {
	isTapTarget bool
	isOutgoing  bool
}
