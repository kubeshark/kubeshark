package source

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type pcapHandle struct {
	source  *gopacket.PacketSource
	capture *pcap.Handle
}

func (h *pcapHandle) NextPacket() (packet gopacket.Packet, err error) {
	return h.source.NextPacket()
}
func (h *pcapHandle) SetDecoder(decoder gopacket.Decoder, lazy bool, noCopy bool) {
	h.source = gopacket.NewPacketSource(h.capture, decoder)
	h.source.Lazy = lazy
	h.source.NoCopy = noCopy
}

func (h *pcapHandle) SetBPF(expr string) (err error) {
	return h.capture.SetBPFFilter(expr)
}

func (h *pcapHandle) LinkType() layers.LinkType {
	return h.capture.LinkType()
}

func (h *pcapHandle) Stats() (packetsReceived uint, packetsDropped uint, err error) {
	var stats *pcap.Stats
	stats, err = h.capture.Stats()
	packetsReceived = uint(stats.PacketsReceived)
	packetsDropped = uint(stats.PacketsDropped)
	return
}

func (h *pcapHandle) Close() (err error) {
	h.capture.Close()
	return
}

func newPcapHandle(filename string, device string, snaplen int, promisc bool, tstype string) (handle Handle, err error) {
	var capture *pcap.Handle

	if filename != "" {
		if capture, err = pcap.OpenOffline(filename); err != nil {
			err = fmt.Errorf("PCAP OpenOffline error: %v", err)
			return
		}
	} else {
		// This is a little complicated because we want to allow all possible options
		// for creating the packet capture handle... instead of all this you can
		// just call pcap.OpenLive if you want a simple handle.
		var inactive *pcap.InactiveHandle
		inactive, err = pcap.NewInactiveHandle(device)
		if err != nil {
			err = fmt.Errorf("could not create: %v", err)
			return
		}
		defer inactive.CleanUp()
		if err = inactive.SetSnapLen(snaplen); err != nil {
			err = fmt.Errorf("could not set snap length: %v", err)
			return
		} else if err = inactive.SetPromisc(promisc); err != nil {
			err = fmt.Errorf("could not set promisc mode: %v", err)
			return
		} else if err = inactive.SetTimeout(time.Second); err != nil {
			err = fmt.Errorf("could not set timeout: %v", err)
			return
		}
		if tstype != "" {
			var t pcap.TimestampSource
			if t, err = pcap.TimestampSourceFromString(tstype); err != nil {
				err = fmt.Errorf("supported timestamp types: %v", inactive.SupportedTimestamps())
				return
			} else if err = inactive.SetTimestampSource(t); err != nil {
				err = fmt.Errorf("supported timestamp types: %v", inactive.SupportedTimestamps())
				return
			}
		}
		if capture, err = inactive.Activate(); err != nil {
			err = fmt.Errorf("PCAP Activate error: %v", err)
			return
		}
	}

	handle = &pcapHandle{
		capture: capture,
	}

	return
}
