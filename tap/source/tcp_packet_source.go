package source

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/ip4defrag"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
	"github.com/up9inc/mizu/tap/diagnose"
	"golang.org/x/net/bpf"
)

type tcpPacketSource struct {
	source    gopacket.ZeroCopyPacketDataSource
	Handle    *afpacket.TPacket
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

func newAfpacketHandle(device string, snaplen int, block_size int, num_blocks int,
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

	fmt.Printf("[afpacketComputeSize] targetSizeMb: %v snaplen: %v pageSize: %v\n", targetSizeMb, snaplen, pageSize)

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

func newTcpPacketSource(name, filename string, interfaceName string,
	behaviour TcpPacketSourceBehaviour, origin api.Capture) (*tcpPacketSource, error) {
	var err error

	result := &tcpPacketSource{
		name:      name,
		defragger: ip4defrag.NewIPv4Defragmenter(),
		Behaviour: &behaviour,
		Origin:    origin,
	}

	szFrame, szBlock, numBlocks, err := afpacketComputeSize(8, 65535, os.Getpagesize())
	if err != nil {
		panic(err)
	}
	result.Handle, err = newAfpacketHandle(interfaceName, szFrame, szBlock, numBlocks, false, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	if behaviour.BpfFilter != "" {
		logger.Log.Infof("Using BPF filter %q", behaviour.BpfFilter)
		if err = result.setBPFFilter(behaviour.BpfFilter); err != nil {
			return nil, fmt.Errorf("BPF filter error: %v", err)
		}
	}

	result.source = gopacket.ZeroCopyPacketDataSource(result.Handle)

	return result, nil
}

func (source *tcpPacketSource) String() string {
	return source.name
}

func (source *tcpPacketSource) setBPFFilter(expr string) (err error) {
	pcapBPF, err := pcap.CompileBPFFilter(layers.LinkTypeEthernet, 65535, expr)
	if err != nil {
		panic(err)
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
	return source.Handle.SetBPF(bpfIns)
}

func (source *tcpPacketSource) close() {
	if source.Handle != nil {
		source.Handle.Close()
	}
}

func (source *tcpPacketSource) readPackets(ipdefrag bool, packets chan<- TcpPacketInfo) {
	logger.Log.Infof("Start reading packets from %v", source.name)

	for {
		data, ci, err := source.source.ZeroCopyReadPacketData()

		decoder := gopacket.DecodersByLayerName[fmt.Sprintf("%s", layers.LinkTypeEthernet)]
		decodeOptions := gopacket.DecodeOptions{Lazy: false, NoCopy: true}

		packet := gopacket.NewPacket(data, decoder, decodeOptions)
		m := packet.Metadata()
		m.CaptureInfo = ci
		m.Truncated = m.Truncated || ci.CaptureLength < ci.Length

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
