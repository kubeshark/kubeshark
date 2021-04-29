package main

import (
	"fmt"
	"sync"

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
	wg        sync.WaitGroup
	doHTTP    bool
	harWriter *HarWriter
}

func (factory *tcpStreamFactory) New(net, transport gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	Debug("* NEW: %s %s\n", net, transport)
	fsmOptions := reassembly.TCPSimpleFSMOptions{
		SupportMissingEstablishment: *allowmissinginit,
	}
	Debug("Current App Ports: %v\n", appPorts)
	dstIp := net.Dst().String()
	dstPort := int(tcp.DstPort)

	if factory.shouldNotifyOnOutboundLink(dstIp, dstPort) {
		broadcastOutboundLink(net.Src().String(), dstIp, dstPort)
	}
	isHTTP := factory.shouldTap(dstIp, dstPort)
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
	if stream.isHTTP {
		stream.client = httpReader{
			msgQueue: make(chan httpReaderDataMsg),
			ident:    fmt.Sprintf("%s %s", net, transport),
			tcpID: tcpID{
				srcIP:   net.Src().String(),
				dstIP:   net.Dst().String(),
				srcPort: transport.Src().String(),
				dstPort: transport.Dst().String(),
			},
			hexdump:  *hexdump,
			parent:   stream,
			isClient: true,
			harWriter: factory.harWriter,
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
			hexdump: *hexdump,
			parent:  stream,
			harWriter: factory.harWriter,
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

func (factory *tcpStreamFactory) shouldTap(dstIP string, dstPort int) bool {
	return true  // TODO: this is only for checking it now

	if hostMode {
		return inArrayString(hostAppAddresses, fmt.Sprintf("%s:%d", dstIP, dstPort))
	} else {
		isTappedPort := dstPort == 80 || (appPorts != nil && (inArrayInt(appPorts, dstPort)))
		if !isTappedPort {
			return false
		}

		if !*anydirection {
			isDirectedHere := inArrayString(ownIps, dstIP)
			if !isDirectedHere {
				return false
			}
		}

		return true
	}
}

func (factory *tcpStreamFactory) shouldNotifyOnOutboundLink(dstIP string, dstPort int) bool {
	if inArrayInt(remoteOnlyOutboundPorts, dstPort) {
		isDirectedHere := inArrayString(ownIps, dstIP)
		return !isDirectedHere && !isPrivateIP(dstIP)
	}
	return true
}
