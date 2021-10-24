package tap

import (
	"fmt"
	"io"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/ip4defrag"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/diagnose"
)

type tcpPacketSource struct {
	source    *gopacket.PacketSource
	handle    *pcap.Handle
	defragger *ip4defrag.IPv4Defragmenter
	behaviour *tcpPacketSourceBehaviour
}

type tcpPacketSourceBehaviour struct {
	snapLength  int
	promisc     bool
	tstype      string
	decoderName string
	lazy        bool
	bpfFilter   string
}

type tcpPacketInfo struct {
	packet gopacket.Packet
	source *tcpPacketSource
}

func NewTcpPacketSource(filename string, interfaceName string,
	behaviour tcpPacketSourceBehaviour) (*tcpPacketSource, error) {
	var err error

	result := &tcpPacketSource{
		defragger: ip4defrag.NewIPv4Defragmenter(),
		behaviour: &behaviour,
	}

	if filename != "" {
		if result.handle, err = pcap.OpenOffline(filename); err != nil {
			return result, fmt.Errorf("PCAP OpenOffline error: %v", err)
		}
	} else {
		// This is a little complicated because we want to allow all possible options
		// for creating the packet capture handle... instead of all this you can
		// just call pcap.OpenLive if you want a simple handle.
		inactive, err := pcap.NewInactiveHandle(interfaceName)
		if err != nil {
			return result, fmt.Errorf("could not create: %v", err)
		}
		defer inactive.CleanUp()
		if err = inactive.SetSnapLen(behaviour.snapLength); err != nil {
			return result, fmt.Errorf("could not set snap length: %v", err)
		} else if err = inactive.SetPromisc(behaviour.promisc); err != nil {
			return result, fmt.Errorf("could not set promisc mode: %v", err)
		} else if err = inactive.SetTimeout(time.Second); err != nil {
			return result, fmt.Errorf("could not set timeout: %v", err)
		}
		if behaviour.tstype != "" {
			if t, err := pcap.TimestampSourceFromString(behaviour.tstype); err != nil {
				return result, fmt.Errorf("supported timestamp types: %v", inactive.SupportedTimestamps())
			} else if err := inactive.SetTimestampSource(t); err != nil {
				return result, fmt.Errorf("supported timestamp types: %v", inactive.SupportedTimestamps())
			}
		}
		if result.handle, err = inactive.Activate(); err != nil {
			return result, fmt.Errorf("PCAP Activate error: %v", err)
		}
	}
	if behaviour.bpfFilter != "" {
		logger.Log.Infof("Using BPF filter %q", behaviour.bpfFilter)
		if err = result.handle.SetBPFFilter(behaviour.bpfFilter); err != nil {
			return nil, fmt.Errorf("BPF filter error: %v", err)
		}
	}

	var dec gopacket.Decoder
	var ok bool
	if behaviour.decoderName == "" {
		behaviour.decoderName = result.handle.LinkType().String()
	}
	if dec, ok = gopacket.DecodersByLayerName[behaviour.decoderName]; !ok {
		return nil, fmt.Errorf("no decoder named %v", behaviour.decoderName)
	}
	result.source = gopacket.NewPacketSource(result.handle, dec)
	result.source.Lazy = behaviour.lazy
	result.source.NoCopy = true

	return result, nil
}

func (source *tcpPacketSource) close() {
	if source.handle != nil {
		source.handle.Close()
	}
}

func (source *tcpPacketSource) readPackets(ipdefrag bool, packets chan<- tcpPacketInfo) error {
	for {
		packet, err := source.source.NextPacket()

		if err == io.EOF {
			return err
		} else if err != nil {
			if err.Error() != "Timeout Expired" {
				logger.Log.Debugf("Error: %T", err)
			}
			continue
		}

		// defrag the IPv4 packet if required
		if !ipdefrag {
			ip4Layer := packet.Layer(layers.LayerTypeIPv4)
			if ip4Layer == nil {
				continue
			}
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

		packets <- tcpPacketInfo{
			packet: packet,
			source: source,
		}
	}
}
