package tap

import (
	"fmt"
	"sync"

	"github.com/romana/rlog"

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
	doHTTP             bool
	harWriter          *HarWriter
	outbountLinkWriter *OutboundLinkWriter
}

func (factory *tcpStreamFactory) New(net, transport gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	rlog.Debugf("* NEW: %s %s", net, transport)
	fsmOptions := reassembly.TCPSimpleFSMOptions{
		SupportMissingEstablishment: *allowmissinginit,
	}
	rlog.Debugf("Current App Ports: %v", gSettings.filterPorts)
	srcIp := net.Src().String()
	dstIp := net.Dst().String()
	dstPort := int(tcp.DstPort)

	if factory.shouldNotifyOnOutboundLink(dstIp, dstPort) {
		factory.outbountLinkWriter.WriteOutboundLink(net.Src().String(), dstIp, dstPort)
	}
	props := factory.getStreamProps(srcIp, dstIp, dstPort)
	isHTTP := props.isTapTarget
	stream := &tcpStream{
		net:        net,
		transport:  transport,
		isDNS:      tcp.SrcPort == 53 || tcp.DstPort == 53,
		isHTTP:     isHTTP && factory.doHTTP,
		isAMQP:     tcp.SrcPort == 5672 || tcp.DstPort == 5672,
		reversed:   tcp.SrcPort == 80,
		tcpstate:   reassembly.NewTCPSimpleFSM(fsmOptions),
		ident:      fmt.Sprintf("%s:%s", net, transport),
		optchecker: reassembly.NewTCPOptionCheck(),
	}
	if stream.isAMQP {
		x := amqpReaderIO{
			msgQueue: make(chan httpReaderDataMsg),
		}
		factory.wg.Add(1)
		go x.run(&factory.wg)
	} else if stream.isHTTP {
		stream.client = httpReader{
			msgQueue: make(chan httpReaderDataMsg),
			ident:    fmt.Sprintf("%s %s", net, transport),
			tcpID: tcpID{
				srcIP:   net.Src().String(),
				dstIP:   net.Dst().String(),
				srcPort: transport.Src().String(),
				dstPort: transport.Dst().String(),
			},
			hexdump:    *hexdump,
			parent:     stream,
			isClient:   true,
			isOutgoing: props.isOutgoing,
			harWriter:  factory.harWriter,
		}
		stream.server = httpReader{
			msgQueue: make(chan httpReaderDataMsg),
			ident:    fmt.Sprintf("%s %s", net.Reverse(), transport.Reverse()),
			tcpID: tcpID{
				srcIP:   net.Dst().String(),
				dstIP:   net.Src().String(),
				srcPort: transport.Dst().String(),
				dstPort: transport.Src().String(),
			},
			hexdump:    *hexdump,
			parent:     stream,
			isOutgoing: props.isOutgoing,
			harWriter:  factory.harWriter,
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

func (factory *tcpStreamFactory) getStreamProps(srcIP string, dstIP string, dstPort int) *streamProps {
	if hostMode {
		if inArrayString(gSettings.filterAuthorities, fmt.Sprintf("%s:%d", dstIP, dstPort)) == true {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ host1 %s:%d", dstIP, dstPort))
			return &streamProps{isTapTarget: true, isOutgoing: false}
		} else if inArrayString(gSettings.filterAuthorities, dstIP) == true {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ host2 %s", dstIP))
			return &streamProps{isTapTarget: true, isOutgoing: false}
		} else if *anydirection && inArrayString(gSettings.filterAuthorities, srcIP) == true {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ host3 %s", srcIP))
			return &streamProps{isTapTarget: true, isOutgoing: true}
		}
		return &streamProps{isTapTarget: false}
	} else {
		isTappedPort := dstPort == 80 || (gSettings.filterPorts != nil && (inArrayInt(gSettings.filterPorts, dstPort)))
		if !isTappedPort {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("- notHost1 %d", dstPort))
			return &streamProps{isTapTarget: false, isOutgoing: false}
		}

		isOutgoing := !inArrayString(ownIps, dstIP)

		if !*anydirection && isOutgoing {
			rlog.Debugf("getStreamProps %s", fmt.Sprintf("- notHost2"))
			return &streamProps{isTapTarget: false, isOutgoing: isOutgoing}
		}

		rlog.Debugf("getStreamProps %s", fmt.Sprintf("+ notHost3 %s -> %s:%d", srcIP, dstIP, dstPort))
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
