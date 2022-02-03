package tap

import (
	"fmt"
	"sync"
	"time"

	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
	v1 "k8s.io/api/core/v1"

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
	Emitter            api.Emitter
	streamsMap         *tcpStreamMap
	ownIps             []string
	opts               *TapOpts
}

type tcpStreamWrapper struct {
	stream    *tcpStream
	createdAt time.Time
}

// var totalStreams int64 = 0

func NewTcpStreamFactory(emitter api.Emitter, streamsMap *tcpStreamMap, opts *TapOpts) *tcpStreamFactory {
	var ownIps []string

	if localhostIPs, err := getLocalhostIPs(); err != nil {
		// TODO: think this over
		logger.Log.Info("Failed to get self IP addresses")
		logger.Log.Errorf("Getting-Self-Address", "Error getting self ip address: %s (%v,%+v)", err, err, err)
		ownIps = make([]string, 0)
	} else {
		ownIps = localhostIPs
	}

	return &tcpStreamFactory{
		Emitter:    emitter,
		streamsMap: streamsMap,
		ownIps:     ownIps,
		opts:       opts,
	}
}

func (factory *tcpStreamFactory) New(net, transport gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	fsmOptions := reassembly.TCPSimpleFSMOptions{
		SupportMissingEstablishment: *allowmissinginit,
	}
	srcIp := net.Src().String()
	dstIp := net.Dst().String()
	srcPort := transport.Src().String()
	dstPort := transport.Dst().String()

	// totalStreams++
	// fmt.Printf("totalStreams: %v\n", totalStreams)

	// if factory.shouldNotifyOnOutboundLink(dstIp, dstPort) {
	// 	factory.outboundLinkWriter.WriteOutboundLink(net.Src().String(), dstIp, dstPort, "", "")
	// }
	props := factory.getStreamProps(srcIp, srcPort, dstIp, dstPort)
	isTapTarget := props.isTapTarget
	stream := &tcpStream{
		net:             net,
		transport:       transport,
		isDNS:           tcp.SrcPort == 53 || tcp.DstPort == 53,
		isTapTarget:     isTapTarget,
		tcpstate:        reassembly.NewTCPSimpleFSM(fsmOptions),
		ident:           fmt.Sprintf("%s:%s", net, transport),
		optchecker:      reassembly.NewTCPOptionCheck(),
		superIdentifier: &api.SuperIdentifier{},
		streamsMap:      factory.streamsMap,
	}
	if stream.isTapTarget {
		stream.id = factory.streamsMap.nextId()
		for i, extension := range extensions {
			counterPair := &api.CounterPair{
				Request:  0,
				Response: 0,
			}
			stream.clients = append(stream.clients, tcpReader{
				msgQueue:   make(chan tcpReaderDataMsg),
				superTimer: &api.SuperTimer{},
				ident:      fmt.Sprintf("%s %s", net, transport),
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
				extension:          extension,
				emitter:            factory.Emitter,
				counterPair:        counterPair,
			})
			stream.servers = append(stream.servers, tcpReader{
				msgQueue:   make(chan tcpReaderDataMsg),
				superTimer: &api.SuperTimer{},
				ident:      fmt.Sprintf("%s %s", net, transport),
				tcpID: &api.TcpID{
					SrcIP:   net.Dst().String(),
					DstIP:   net.Src().String(),
					SrcPort: transport.Dst().String(),
					DstPort: transport.Src().String(),
				},
				parent:             stream,
				isClient:           false,
				isOutgoing:         props.isOutgoing,
				outboundLinkWriter: factory.outboundLinkWriter,
				extension:          extension,
				emitter:            factory.Emitter,
				counterPair:        counterPair,
			})

			factory.streamsMap.Store(stream.id, &tcpStreamWrapper{
				stream:    stream,
				createdAt: time.Now(),
			})

			factory.wg.Add(2)
			// Start reading from channel stream.reader.bytes
			go stream.clients[i].run(&factory.wg)
			go stream.servers[i].run(&factory.wg)
		}
	}
	return stream
}

func (factory *tcpStreamFactory) WaitGoRoutines() {
	factory.wg.Wait()
}

func inArrayPod(pods []v1.Pod, address string) bool {
	for _, pod := range pods {
		if pod.Status.PodIP == address {
			return true
		}
	}
	return false
}

func (factory *tcpStreamFactory) getStreamProps(srcIP string, srcPort string, dstIP string, dstPort string) *streamProps {
	if factory.opts.HostMode {
		if inArrayPod(tapTargets, fmt.Sprintf("%s:%s", dstIP, dstPort)) {
			return &streamProps{isTapTarget: true, isOutgoing: false}
		} else if inArrayPod(tapTargets, dstIP) {
			return &streamProps{isTapTarget: true, isOutgoing: false}
		} else if inArrayPod(tapTargets, fmt.Sprintf("%s:%s", srcIP, srcPort)) {
			return &streamProps{isTapTarget: true, isOutgoing: true}
		} else if inArrayPod(tapTargets, srcIP) {
			return &streamProps{isTapTarget: true, isOutgoing: true}
		}
		return &streamProps{isTapTarget: false, isOutgoing: false}
	} else {
		return &streamProps{isTapTarget: true}
	}
}

//lint:ignore U1000 will be used in the future
func (factory *tcpStreamFactory) shouldNotifyOnOutboundLink(dstIP string, dstPort int) bool {
	if inArrayInt(remoteOnlyOutboundPorts, dstPort) {
		isDirectedHere := inArrayString(factory.ownIps, dstIP)
		return !isDirectedHere && !isPrivateIP(dstIP)
	}
	return true
}

type streamProps struct {
	isTapTarget bool
	isOutgoing  bool
}
