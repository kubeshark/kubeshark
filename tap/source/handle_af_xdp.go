package source

import (
	"fmt"
	"net"

	"github.com/asavie/xdp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/up9inc/mizu/tap/tlstapper"
	ebpf "github.com/up9inc/mizu/tap/xdp"
)

type afXdpHandle struct {
	ifindex int
	queueId int
	program *xdp.Program
	xsk     *xdp.Socket
}

func (h *afXdpHandle) NextPacket() (packet gopacket.Packet, err error) {
	// If there are any free slots on the Fill queue...
	if n := h.xsk.NumFreeFillSlots(); n > 0 {
		// ...then fetch up to that number of not-in-use
		// descriptors and push them onto the Fill ring queue
		// for the kernel to fill them with the received
		// frames.
		h.xsk.Fill(h.xsk.GetDescs(n))
	}

	// Wait for receive - meaning the kernel has
	// produced one or more descriptors filled with a received
	// frame onto the Rx ring queue.
	var numRx int
	numRx, _, err = h.xsk.Poll(-1)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	if numRx > 0 {
		// Consume the descriptors filled with received frames
		// from the Rx ring queue.
		rxDescs := h.xsk.Receive(numRx)

		// Print the received frames and also modify them
		// in-place replacing the destination MAC address with
		// broadcast address.
		for i := 0; i < len(rxDescs); i++ {
			packetData := h.xsk.GetFrame(rxDescs[i])
			packet = gopacket.NewPacket(packetData, layers.LayerTypeEthernet, gopacket.Default)
		}
	}
	return
}

func (h *afXdpHandle) SetDecoder(decoder gopacket.Decoder, lazy bool, noCopy bool) {
	// TODO: Implement?
}

func (h *afXdpHandle) SetBPF(expr string) (err error) {
	// TODO: Implement?
	return
}

func (h *afXdpHandle) LinkType() layers.LinkType {
	return layers.LinkTypeEthernet
}

func (h *afXdpHandle) Stats() (packetsReceived uint, packetsDropped uint, err error) {
	// TODO: Implement?
	return
}

func (h *afXdpHandle) Close() (err error) {
	h.program.Close()
	err = h.program.Detach(h.ifindex)
	if err != nil {
		return
	}
	err = h.program.Unregister(h.queueId)
	return
}

func newAfXdpHandle(device string) (handle Handle, err error) {
	err = tlstapper.SetupRLimit()
	if err != nil {
		return
	}

	var queueId int = 0
	const protocol uint8 = 4 // IPv4 - https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml

	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	ifindex := -1
	for _, iface := range interfaces {
		if iface.Name == device {
			ifindex = iface.Index
			break
		}
	}
	if ifindex == -1 {
		err = fmt.Errorf("Interface index is -1")
		return
	}

	var program *xdp.Program
	// Create a new XDP eBPF program and attach it to our chosen network link.
	program, err = ebpf.NewIPProtoProgram(protocol, nil)
	if err != nil {
		return
	}
	if err = program.Attach(ifindex); err != nil {
		return
	}

	// Create and initialize an XDP socket attached to our chosen network
	// link.
	var xsk *xdp.Socket
	xsk, err = xdp.NewSocket(ifindex, queueId, nil)
	if err != nil {
		return
	}

	// Register our XDP socket file descriptor with the eBPF program so it can be redirected packets
	if err = program.Register(queueId, xsk.FD()); err != nil {
		return
	}

	handle = &afXdpHandle{
		ifindex: ifindex,
		queueId: queueId,
		program: program,
		xsk:     xsk,
	}
	return
}
