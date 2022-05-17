package source

import (
	"fmt"
	"io"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/ip4defrag"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/dbgctl"
	"github.com/up9inc/mizu/tap/diagnose"
)

type tcpPacketSource struct {
	source    *gopacket.PacketSource
	handle    *pcap.Handle
	defragger *ip4defrag.IPv4Defragmenter
	Behaviour *TcpPacketSourceBehaviour
	name      string
	Origin    api.Capture
}

type TcpPacketSourceBehaviour struct {
	SnapLength  int
	Promisc     bool
	Tstype      string
	DecoderName string
	Lazy        bool
	BpfFilter   string
}

type TcpPacketInfo struct {
	Packet gopacket.Packet
	Source *tcpPacketSource
}

func newTcpPacketSource(name, filename string, interfaceName string,
	behaviour TcpPacketSourceBehaviour, origin api.Capture) (*tcpPacketSource, error) {
	var err error

	result := &tcpPacketSource{
		name:      name,
		defragger: ip4defrag.NewIPv4Defragmenter(),
		Behaviour: &behaviour,
		Origin:    origin,
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
		if err = inactive.SetSnapLen(behaviour.SnapLength); err != nil {
			return result, fmt.Errorf("could not set snap length: %v", err)
		} else if err = inactive.SetPromisc(behaviour.Promisc); err != nil {
			return result, fmt.Errorf("could not set promisc mode: %v", err)
		} else if err = inactive.SetTimeout(time.Second); err != nil {
			return result, fmt.Errorf("could not set timeout: %v", err)
		}
		if behaviour.Tstype != "" {
			if t, err := pcap.TimestampSourceFromString(behaviour.Tstype); err != nil {
				return result, fmt.Errorf("supported timestamp types: %v", inactive.SupportedTimestamps())
			} else if err := inactive.SetTimestampSource(t); err != nil {
				return result, fmt.Errorf("supported timestamp types: %v", inactive.SupportedTimestamps())
			}
		}
		if result.handle, err = inactive.Activate(); err != nil {
			return result, fmt.Errorf("PCAP Activate error: %v", err)
		}
	}
	if behaviour.BpfFilter != "" {
		logger.Log.Infof("Using BPF filter %q", behaviour.BpfFilter)
		if err = result.handle.SetBPFFilter(behaviour.BpfFilter); err != nil {
			return nil, fmt.Errorf("BPF filter error: %v", err)
		}
	}

	var dec gopacket.Decoder
	var ok bool
	if behaviour.DecoderName == "" {
		behaviour.DecoderName = result.handle.LinkType().String()
	}
	if dec, ok = gopacket.DecodersByLayerName[behaviour.DecoderName]; !ok {
		return nil, fmt.Errorf("no decoder named %v", behaviour.DecoderName)
	}
	result.source = gopacket.NewPacketSource(result.handle, dec)
	result.source.Lazy = behaviour.Lazy
	result.source.NoCopy = true

	return result, nil
}

func (source *tcpPacketSource) String() string {
	return source.name
}

func (source *tcpPacketSource) setBPFFilter(expr string) (err error) {
	return source.handle.SetBPFFilter(expr)
}

func (source *tcpPacketSource) close() {
	if source.handle != nil {
		source.handle.Close()
	}
}

func (source *tcpPacketSource) readPackets(ipdefrag bool, packets chan<- TcpPacketInfo) {
	if dbgctl.MizuTapperDisablePcap {
		return
	}
	logger.Log.Infof("Start reading packets from %v", source.name)

	for {
		packet, err := source.source.NextPacket()

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
