package source

import (
	"fmt"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"golang.org/x/net/bpf"
)

type afPacketHandle struct {
	source        gopacket.ZeroCopyPacketDataSource
	capture       *afpacket.TPacket
	decoder       gopacket.Decoder
	decodeOptions gopacket.DecodeOptions
}

func (h *afPacketHandle) NextPacket() (packet gopacket.Packet, err error) {
	var data []byte
	var ci gopacket.CaptureInfo
	data, ci, err = h.source.ZeroCopyReadPacketData()
	if err != nil {
		return
	}

	packet = gopacket.NewPacket(data, h.decoder, h.decodeOptions)
	m := packet.Metadata()
	m.CaptureInfo = ci
	m.Truncated = m.Truncated || ci.CaptureLength < ci.Length
	return
}

func (h *afPacketHandle) SetDecoder(decoder gopacket.Decoder, lazy bool, noCopy bool) {
	h.decoder = decoder
	h.decodeOptions = gopacket.DecodeOptions{Lazy: lazy, NoCopy: noCopy}
}

func (h *afPacketHandle) SetBPF(expr string) (err error) {
	var pcapBPF []pcap.BPFInstruction
	pcapBPF, err = pcap.CompileBPFFilter(layers.LinkTypeEthernet, 65535, expr)
	if err != nil {
		return
	}
	bpfIns := []bpf.RawInstruction{}
	for _, ins := range pcapBPF {
		bpfIns2 := bpf.RawInstruction{
			Op: ins.Code,
			Jt: ins.Jt,
			Jf: ins.Jf,
			K:  ins.K,
		}
		bpfIns = append(bpfIns, bpfIns2)
	}
	err = h.capture.SetBPF(bpfIns)
	return
}

func (h *afPacketHandle) LinkType() layers.LinkType {
	return layers.LinkTypeEthernet
}

func (h *afPacketHandle) Stats() (packetsReceived uint, packetsDropped uint, err error) {
	var stats afpacket.SocketStatsV3
	_, stats, err = h.capture.SocketStats()
	packetsReceived = stats.Packets()
	packetsDropped = stats.Drops()
	return
}

func (h *afPacketHandle) Close() (err error) {
	h.capture.Close()
	return
}

func newAfpacketHandle(device string, targetSizeMb int, snaplen int) (handle Handle, err error) {
	snaplen -= 1
	if snaplen < 0 {
		snaplen = 0
	}
	szFrame, szBlock, numBlocks, err := afpacketComputeSize(targetSizeMb, snaplen, os.Getpagesize())
	if err != nil {
		return
	}
	var capture *afpacket.TPacket
	capture, err = newAfpacket(device, szFrame, szBlock, numBlocks, false, pcap.BlockForever)
	if err != nil {
		return
	}
	handle = &afPacketHandle{
		capture: capture,
		source:  gopacket.ZeroCopyPacketDataSource(capture),
	}
	return
}

func newAfpacket(device string, snaplen int, block_size int, num_blocks int,
	useVLAN bool, timeout time.Duration) (*afpacket.TPacket, error) {

	var h *afpacket.TPacket
	var err error

	if device == "any" {
		h, err = afpacket.NewTPacket(
			afpacket.OptFrameSize(snaplen),
			afpacket.OptBlockSize(block_size),
			afpacket.OptNumBlocks(num_blocks),
			afpacket.OptAddVLANHeader(useVLAN),
			afpacket.OptPollTimeout(timeout),
			afpacket.SocketRaw,
			afpacket.TPacketVersion3)
	} else {
		h, err = afpacket.NewTPacket(
			afpacket.OptInterface(device),
			afpacket.OptFrameSize(snaplen),
			afpacket.OptBlockSize(block_size),
			afpacket.OptNumBlocks(num_blocks),
			afpacket.OptAddVLANHeader(useVLAN),
			afpacket.OptPollTimeout(timeout),
			afpacket.SocketRaw,
			afpacket.TPacketVersion3)
	}
	return h, err
}

// afpacketComputeSize computes the block_size and the num_blocks in such a way that the
// allocated mmap buffer is close to but smaller than target_size_mb.
// The restriction is that the block_size must be divisible by both the
// frame size and page size.
func afpacketComputeSize(targetSizeMb int, snaplen int, pageSize int) (
	frameSize int, blockSize int, numBlocks int, err error) {

	if snaplen < pageSize {
		frameSize = pageSize / (pageSize / snaplen)
	} else {
		frameSize = (snaplen/pageSize + 1) * pageSize
	}

	// 128 is the default from the gopacket library so just use that
	blockSize = frameSize * 128
	numBlocks = (targetSizeMb * 1024 * 1024) / blockSize

	if numBlocks == 0 {
		return 0, 0, 0, fmt.Errorf("Interface buffersize is too small")
	}

	return frameSize, blockSize, numBlocks, nil
}
