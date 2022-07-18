package source

import (
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/asavie/xdp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/tlstapper"
	ebpf "github.com/up9inc/mizu/tap/xdp"
)

type afXdpHandle struct {
	ifindexes       []int
	queueId         int
	program         *xdp.Program
	xsks            []*xdp.Socket
	ngInterfaces    []pcapgo.NgInterface
	ngReader        *pcapgo.NgReader
	ngWriter        *pcapgo.NgWriter
	decoder         gopacket.Decoder
	decodeOptions   gopacket.DecodeOptions
	pipeReader      *io.PipeReader
	packetsReceived uint
}

func (h *afXdpHandle) NextPacket() (packet gopacket.Packet, err error) {
	if h.ngReader == nil {
		h.ngReader, err = pcapgo.NewNgReader(h.pipeReader, pcapgo.NgReaderOptions{})
		if err != nil {
			return
		}
	}

	var data []byte
	var ci gopacket.CaptureInfo
	data, ci, err = h.ngReader.ZeroCopyReadPacketData()
	if err != nil {
		return
	}
	h.packetsReceived++

	packet = gopacket.NewPacket(data, h.decoder, h.decodeOptions)
	m := packet.Metadata()
	m.CaptureInfo = ci
	m.Truncated = m.Truncated || ci.CaptureLength < ci.Length
	return
}

func (h *afXdpHandle) SetDecoder(decoder gopacket.Decoder, lazy bool, noCopy bool) {
	h.decoder = decoder
	h.decodeOptions = gopacket.DecodeOptions{Lazy: lazy, NoCopy: noCopy}
}

func (h *afXdpHandle) SetBPF(expr string) (err error) {
	for _, ngInterface := range h.ngInterfaces {
		ngInterface.Filter = expr
	}
	return
}

func (h *afXdpHandle) LinkType() layers.LinkType {
	return layers.LinkTypeEthernet
}

func (h *afXdpHandle) Stats() (packetsReceived uint, packetsDropped uint, err error) {
	packetsReceived = h.packetsReceived
	return
}

func (h *afXdpHandle) Close() (err error) {
	h.program.Close()
	for _, ifindex := range h.ifindexes {
		err = h.program.Detach(ifindex)
		if err != nil {
			return
		}
	}
	err = h.program.Unregister(h.queueId)
	return
}

func (h *afXdpHandle) poll() {
	for i, xsk := range h.xsks {
		go h.pollSocket(xsk, h.ifindexes[i])
	}
}

func (h *afXdpHandle) pollSocket(xsk *xdp.Socket, ifindex int) {
	var err error
	// If there are any free slots on the Fill queue...
	if n := xsk.NumFreeFillSlots(); n > 0 {
		// ...then fetch up to that number of not-in-use
		// descriptors and push them onto the Fill ring queue
		// for the kernel to fill them with the received
		// frames.
		xsk.Fill(xsk.GetDescs(n))
	}

	// Wait for receive - meaning the kernel has
	// produced one or more descriptors filled with a received
	// frame onto the Rx ring queue.
	var numRx int
	numRx, _, err = xsk.Poll(-1)
	if err != nil {
		logger.Log.Debugf("XDP polling error: %v", err)
		return
	}

	if numRx > 0 {
		// Consume the descriptors filled with received frames
		// from the Rx ring queue.
		rxDescs := xsk.Receive(numRx)

		// Print the received frames and also modify them
		// in-place replacing the destination MAC address with
		// broadcast address.
		for i := 0; i < len(rxDescs); i++ {
			packet := gopacket.NewPacket(xsk.GetFrame(rxDescs[i]), layers.LayerTypeEthernet, gopacket.Default)
			data := packet.Data()
			info := gopacket.CaptureInfo{
				Timestamp:      time.Now(),
				CaptureLength:  len(data),
				Length:         len(data),
				InterfaceIndex: ifindex,
			}
			err = h.ngWriter.WritePacket(info, data)
			if err != nil {
				logger.Log.Debugf("XDP error writing packet to pipe: %v", err)
			}
			h.ngWriter.Flush()
		}
	}
}

func (h *afXdpHandle) initPcapPipe() (err error) {
	var pipeWriter *io.PipeWriter
	h.pipeReader, pipeWriter = io.Pipe()

	h.ngWriter, err = pcapgo.NewNgWriterInterface(pipeWriter, h.ngInterfaces[0], pcapgo.NgWriterOptions{})
	if err != nil {
		return
	}

	for _, ifc := range h.ngInterfaces[1:] {
		_, err = h.ngWriter.AddInterface(ifc)
		if err != nil {
			return
		}
	}

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

	var ifindexes []int
	var ngInterfaces []pcapgo.NgInterface
	for _, iface := range interfaces {
		if iface.Name == device || device == "any" {
			ifindexes = append(ifindexes, iface.Index)
			if device != "any" {
				break
			}
		}

		ngInterfaces = append(ngInterfaces, pcapgo.NgInterface{
			Name:       iface.Name,
			Comment:    "XDP action",
			LinkType:   layers.LinkTypeEthernet,
			SnapLength: uint32(math.MaxUint16),
		})
	}

	if len(ifindexes) == 0 {
		err = fmt.Errorf("Interface indexes is empty!")
		return
	}

	var program *xdp.Program
	// Create a new XDP eBPF program and attach it to our chosen network link.
	if protocol == 0 {
		program, err = xdp.NewProgram(queueId + 1)
	} else {
		program, err = ebpf.NewIPProtoProgram(protocol, nil)
	}
	if err != nil {
		return
	}

	var xsks []*xdp.Socket
	for _, ifindex := range ifindexes {
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
		xsks = append(xsks, xsk)

		// Register our XDP socket file descriptor with the eBPF program so it can be redirected packets
		if err = program.Register(queueId, xsk.FD()); err != nil {
			return
		}
	}

	h := &afXdpHandle{
		ifindexes:    ifindexes,
		queueId:      queueId,
		program:      program,
		xsks:         xsks,
		ngInterfaces: ngInterfaces,
	}

	err = h.initPcapPipe()
	if err != nil {
		return
	}

	h.poll()

	handle = h

	return
}
