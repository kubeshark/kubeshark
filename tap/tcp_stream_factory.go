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
	outbountLinkWriter *OutboundLinkWriter
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
	// 	factory.outbountLinkWriter.WriteOutboundLink(net.Src().String(), dstIp, dstPort, "", "")
	// }
	props := factory.getStreamProps(srcIp, dstIp, srcPort, dstPort, factory.AllExtensionPorts)
	isTapTarget := props.isTapTarget
	stream := &tcpStream{
		net:         net,
		transport:   transport,
		isDNS:       tcp.SrcPort == 53 || tcp.DstPort == 53,
		isTapTarget: isTapTarget,
		reversed:    props.reversed,
		tcpstate:    reassembly.NewTCPSimpleFSM(fsmOptions),
		ident:       fmt.Sprintf("%s:%s", net, transport),
		optchecker:  reassembly.NewTCPOptionCheck(),
	}
	if stream.isTapTarget {
		stream.client = tcpReader{
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
			outboundLinkWriter: factory.outbountLinkWriter,
			Emitter:            factory.Emitter,
		}
		stream.server = tcpReader{
			msgQueue: make(chan tcpReaderDataMsg),
			ident:    fmt.Sprintf("%s %s", net.Reverse(), transport.Reverse()),
			tcpID: &api.TcpID{
				SrcIP:   net.Dst().String(),
				DstIP:   net.Src().String(),
				SrcPort: transport.Dst().String(),
				DstPort: transport.Src().String(),
			},
			parent:             stream,
			isOutgoing:         props.isOutgoing,
			outboundLinkWriter: factory.outbountLinkWriter,
			Emitter:            factory.Emitter,
		}
		factory.wg.Add(2)
		// Start reading from channels stream.client.bytes and stream.server.bytes
		go stream.client.run(&factory.wg)
		go stream.server.run(&factory.wg)
	}
	return stream
}

func (factory *tcpStreamFactory) WaitGoRoutines() {
	factory.wg.Wait()
}

func (factory *tcpStreamFactory) getStreamProps(srcIP string, dstIP string, srcPort string, dstPort string, allExtensionPorts []string) *streamProps {
	reversed := false
	if hostMode {
		// TODO: Implement reversed for the `hostMode`
		if inArrayString(gSettings.filterAuthorities, fmt.Sprintf("%s:%s", dstIP, dstPort)) == true {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ host1 %s:%s", dstIP, dstPort))
			return &streamProps{isTapTarget: true, isOutgoing: false, reversed: reversed}
		} else if inArrayString(gSettings.filterAuthorities, dstIP) == true {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ host2 %s", dstIP))
			return &streamProps{isTapTarget: true, isOutgoing: false, reversed: reversed}
		} else if *anydirection && inArrayString(gSettings.filterAuthorities, srcIP) == true {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ host3 %s", srcIP))
			return &streamProps{isTapTarget: true, isOutgoing: true, reversed: reversed}
		}
		return &streamProps{isTapTarget: false, reversed: reversed}
	} else {
		// TODO: Bring back `filterPorts` as a string if it's really needed
		// (gSettings.filterPorts != nil && (inArrayInt(gSettings.filterPorts, dstPort)))
		isTappedPort := containsPort(allExtensionPorts, dstPort)
		if !isTappedPort && containsPort(allExtensionPorts, srcPort) {
			isTappedPort = true
			reversed = true
		}
		if !isTappedPort {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("- notHost1 %s", dstPort))
			return &streamProps{isTapTarget: false, isOutgoing: false, reversed: reversed}
		}

		isOutgoing := !inArrayString(ownIps, dstIP)

		if !*anydirection && isOutgoing {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("- notHost2"))
			return &streamProps{isTapTarget: false, isOutgoing: isOutgoing, reversed: reversed}
		}

		rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ notHost3 %s -> %s:%s", srcIP, dstIP, dstPort))
		return &streamProps{isTapTarget: true, reversed: reversed}
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
	reversed    bool
}
