package tap

import (
	"fmt"
	"sync"

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
	wg         sync.WaitGroup
	Emitter    api.Emitter
	streamsMap *api.TcpStreamMap
	ownIps     []string
	opts       *TapOpts
}

func NewTcpStreamFactory(emitter api.Emitter, streamsMap *api.TcpStreamMap, opts *TapOpts) *tcpStreamFactory {
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

	props := factory.getStreamProps(srcIp, srcPort, dstIp, dstPort)
	isTapTarget := props.isTapTarget
	stream := &api.TcpStream{
		Net:             net,
		Transport:       transport,
		IsDNS:           tcp.SrcPort == 53 || tcp.DstPort == 53,
		IsTapTarget:     isTapTarget,
		TcpState:        reassembly.NewTCPSimpleFSM(fsmOptions),
		Ident:           fmt.Sprintf("%s:%s", net, transport),
		Optchecker:      reassembly.NewTCPOptionCheck(),
		ProtoIdentifier: &api.ProtoIdentifier{},
		StreamsMap:      factory.streamsMap,
		Origin:          getPacketOrigin(ac),
	}
	if stream.IsTapTarget {
		stream.Id = factory.streamsMap.NextId()
		for i, extension := range extensions {
			reqResMatcher := extension.Dissector.NewResponseRequestMatcher()
			counterPair := &api.CounterPair{
				Request:  0,
				Response: 0,
			}
			stream.Clients = append(stream.Clients, api.TcpReader{
				MsgQueue:   make(chan api.TcpReaderDataMsg),
				Progress:   &api.ReadProgress{},
				SuperTimer: &api.SuperTimer{},
				Ident:      fmt.Sprintf("%s %s", net, transport),
				TcpID: &api.TcpID{
					SrcIP:   srcIp,
					DstIP:   dstIp,
					SrcPort: srcPort,
					DstPort: dstPort,
				},
				Parent:        stream,
				IsClient:      true,
				IsOutgoing:    props.isOutgoing,
				Extension:     extension,
				Emitter:       factory.Emitter,
				CounterPair:   counterPair,
				ReqResMatcher: reqResMatcher,
			})
			stream.Servers = append(stream.Servers, api.TcpReader{
				MsgQueue:   make(chan api.TcpReaderDataMsg),
				Progress:   &api.ReadProgress{},
				SuperTimer: &api.SuperTimer{},
				Ident:      fmt.Sprintf("%s %s", net, transport),
				TcpID: &api.TcpID{
					SrcIP:   net.Dst().String(),
					DstIP:   net.Src().String(),
					SrcPort: transport.Dst().String(),
					DstPort: transport.Src().String(),
				},
				Parent:        stream,
				IsClient:      false,
				IsOutgoing:    props.isOutgoing,
				Extension:     extension,
				Emitter:       factory.Emitter,
				CounterPair:   counterPair,
				ReqResMatcher: reqResMatcher,
			})

			factory.streamsMap.Store(stream.Id, stream)

			factory.wg.Add(2)
			// Start reading from channel stream.reader.bytes
			go stream.Clients[i].Run(filteringOptions, &factory.wg)
			go stream.Servers[i].Run(filteringOptions, &factory.wg)
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

func getPacketOrigin(ac reassembly.AssemblerContext) api.Capture {
	c, ok := ac.(*context)

	if !ok {
		// If ac is not our context, fallback to Pcap
		return api.Pcap
	}

	return c.Origin
}

type streamProps struct {
	isTapTarget bool
	isOutgoing  bool
}
