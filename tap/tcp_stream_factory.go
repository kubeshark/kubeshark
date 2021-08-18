package tap

import (
	"bufio"
	"fmt"
	"log"

	"github.com/romana/rlog"
	"github.com/up9inc/mizu/tap/api"

	"github.com/google/gopacket" // pulls in all layers decoders
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

type tcpStreamFactory struct {
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

func (h *tcpStream) clientRun(tcpID *api.TcpID) {
	b := bufio.NewReader(&h.r)
	for _, extension := range extensions {
		if containsPort(extension.OutboundPorts, h.transport.Dst().String()) {
			extension.Dissector.Ping()
			extension.Dissector.Dissect(b, true, tcpID)
		}
	}
}

func (h *tcpStream) serverRun(tcpID *api.TcpID) {
	b := bufio.NewReader(&h.r)
	for _, extension := range extensions {
		if containsPort(extension.OutboundPorts, h.transport.Src().String()) {
			extension.Dissector.Ping()
			extension.Dissector.Dissect(b, false, tcpID)
		}
	}
}

func (h *tcpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	log.Printf("* NEW: %s %s\n", net, transport)
	stream := &tcpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
	}
	tcpID := &api.TcpID{
		SrcIP:   net.Src().String(),
		DstIP:   net.Dst().String(),
		SrcPort: transport.Src().String(),
		DstPort: transport.Dst().String(),
	}
	if containsPort(allOutboundPorts, transport.Dst().String()) {
		go stream.clientRun(tcpID)
	} else if containsPort(allOutboundPorts, transport.Src().String()) {
		go stream.serverRun(tcpID)
	}
	return &stream.r
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
