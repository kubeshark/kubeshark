package tap

import (
	"bufio"
	"fmt"
	"io"
	"sync"

	"github.com/romana/rlog"

	"github.com/bradleyfalzon/tlsx"
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
	outbountLinkWriter *OutboundLinkWriter
}

const checkTLSPacketAmount = 100

func containsPort(ports []string, port string) bool {
	for _, x := range ports {
		if x == port {
			return true
		}
	}
	return false
}

func (h *tcpStream) run(wg *sync.WaitGroup) {
	defer wg.Done()
	for _, extension := range extensions {
		if containsPort(extension.Ports, h.transport.Dst().String()) {
			b := bufio.NewReader(h)
			extension.Dissector.Ping()
			extension.Dissector.Dissect(b)
		}
	}
	// b := bufio.NewReader(h)
	// fmt.Printf("b: %v\n", b)
}

func (h *tcpStream) Read(p []byte) (int, error) {
	var msg tcpReaderDataMsg

	ok := true
	for ok && len(h.data) == 0 {
		msg, ok = <-h.msgQueue
		h.data = msg.bytes

		h.captureTime = msg.timestamp
		if len(h.data) > 0 {
			h.packetsSeen += 1
		}
		if h.packetsSeen < checkTLSPacketAmount && len(msg.bytes) > 5 { // packets with less than 5 bytes cause tlsx to panic
			clientHello := tlsx.ClientHello{}
			err := clientHello.Unmarshall(msg.bytes)
			if err == nil {
				// statsTracker.incTlsConnectionsCount()
				fmt.Printf("Detected TLS client hello with SNI %s\n", clientHello.SNI)
				// numericPort, _ := strconv.Atoi(h.tcpID.dstPort)
				// h.outboundLinkWriter.WriteOutboundLink(h.tcpID.srcIP, h.tcpID.dstIP, numericPort, clientHello.SNI, TLSProtocol)
			}
		}
	}
	if !ok || len(h.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, h.data)
	h.data = h.data[l:]
	return l, nil
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

	// if factory.shouldNotifyOnOutboundLink(dstIp, dstPort) {
	// 	factory.outbountLinkWriter.WriteOutboundLink(net.Src().String(), dstIp, dstPort, "", "")
	// }
	props := factory.getStreamProps(srcIp, dstIp, dstPort)
	isHTTP := props.isTapTarget
	stream := &tcpStream{
		net:        net,
		transport:  transport,
		isDNS:      tcp.SrcPort == 53 || tcp.DstPort == 53,
		isHTTP:     isHTTP && factory.doHTTP,
		reversed:   tcp.SrcPort == 80,
		tcpstate:   reassembly.NewTCPSimpleFSM(fsmOptions),
		ident:      fmt.Sprintf("%s:%s", net, transport),
		optchecker: reassembly.NewTCPOptionCheck(),
	}
	factory.wg.Add(1)
	go stream.run(&factory.wg)
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
