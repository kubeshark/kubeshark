package source

import (
	"fmt"
	"io"

	"github.com/google/gopacket"
	"github.com/google/gopacket/ip4defrag"
	"github.com/google/gopacket/layers"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/tap/api"
	"github.com/kubeshark/kubeshark/tap/dbgctl"
	"github.com/kubeshark/kubeshark/tap/diagnose"
)

type Handle interface {
	NextPacket() (packet gopacket.Packet, err error)
	SetDecoder(decoder gopacket.Decoder, lazy bool, noCopy bool)
	SetBPF(expr string) (err error)
	LinkType() layers.LinkType
	Stats() (packetsReceived uint, packetsDropped uint, err error)
	Close() (err error)
}

type tcpPacketSource struct {
	Handle    Handle
	defragger *ip4defrag.IPv4Defragmenter
	Behaviour *TcpPacketSourceBehaviour
	name      string
	Origin    api.Capture
}

type TcpPacketSourceBehaviour struct {
	SnapLength   int
	TargetSizeMb int
	Promisc      bool
	Tstype       string
	DecoderName  string
	Lazy         bool
	BpfFilter    string
}

type TcpPacketInfo struct {
	Packet gopacket.Packet
	Source *tcpPacketSource
}

func newTcpPacketSource(name, filename string, interfaceName string, packetCapture string,
	behaviour TcpPacketSourceBehaviour, origin api.Capture) (*tcpPacketSource, error) {
	var err error

	result := &tcpPacketSource{
		name:      name,
		defragger: ip4defrag.NewIPv4Defragmenter(),
		Behaviour: &behaviour,
		Origin:    origin,
	}

	switch packetCapture {
	case "af_packet":
		result.Handle, err = newAfpacketHandle(
			interfaceName,
			behaviour.TargetSizeMb,
			behaviour.SnapLength,
		)
		if err != nil {
			return nil, err
		}
		logger.Log.Infof("Using AF_PACKET socket as the capture source")
	default:
		result.Handle, err = newPcapHandle(
			filename,
			interfaceName,
			behaviour.SnapLength,
			behaviour.Promisc,
			behaviour.Tstype,
		)
		if err != nil {
			return nil, err
		}
		logger.Log.Infof("Using libpcap as the capture source")
	}

	var decoder gopacket.Decoder
	var ok bool
	if behaviour.DecoderName == "" {
		behaviour.DecoderName = result.Handle.LinkType().String()
	}
	if decoder, ok = gopacket.DecodersByLayerName[behaviour.DecoderName]; !ok {
		return nil, fmt.Errorf("no decoder named %v", behaviour.DecoderName)
	}
	result.Handle.SetDecoder(decoder, behaviour.Lazy, true)

	if behaviour.BpfFilter != "" {
		logger.Log.Infof("Using BPF filter %q", behaviour.BpfFilter)
		if err = result.setBPFFilter(behaviour.BpfFilter); err != nil {
			return nil, fmt.Errorf("BPF filter error: %v", err)
		}
	}

	return result, nil
}

func (source *tcpPacketSource) String() string {
	return source.name
}

func (source *tcpPacketSource) setBPFFilter(expr string) (err error) {
	return source.Handle.SetBPF(expr)
}

func (source *tcpPacketSource) close() {
	if source.Handle != nil {
		source.Handle.Close()
	}
}

func (source *tcpPacketSource) Stats() (packetsReceived uint, packetsDropped uint, err error) {
	return source.Handle.Stats()
}

func (source *tcpPacketSource) readPackets(ipdefrag bool, packets chan<- TcpPacketInfo) {
	if dbgctl.KubesharkTapperDisablePcap {
		return
	}
	logger.Log.Infof("Start reading packets from %v", source.name)

	for {
		packet, err := source.Handle.NextPacket()

		if err == io.EOF {
			logger.Log.Infof("Got EOF while reading packets from %v", source.name)
			return
		} else if err != nil {
			if err.Error() != "Timeout Expired" {
				logger.Log.Debugf("Error while reading from %v - %v", source.name, err)
			}
			continue
		}

		// defrag the IPv4 packet if required
		if ipdefrag {
			if ip4Layer := packet.Layer(layers.LayerTypeIPv4); ip4Layer != nil {
				ip4 := ip4Layer.(*layers.IPv4)
				l := ip4.Length
				newip4, err := source.defragger.DefragIPv4(ip4)
				if err != nil {
					logger.Log.Fatal("Error while de-fragmenting", err)
				} else if newip4 == nil {
					logger.Log.Debugf("Fragment...")
					continue // packet fragment, we don't have whole packet yet.
				}
				if newip4.Length != l {
					diagnose.InternalStats.Ipdefrag++
					logger.Log.Debugf("Decoding re-assembled packet: %s", newip4.NextLayerType())
					pb, ok := packet.(gopacket.PacketBuilder)
					if !ok {
						logger.Log.Panic("Not a PacketBuilder")
					}
					nextDecoder := newip4.NextLayerType()
					_ = nextDecoder.Decode(newip4.Payload, pb)
				}
			}
		}

		packets <- TcpPacketInfo{
			Packet: packet,
			Source: source,
		}
	}
}
