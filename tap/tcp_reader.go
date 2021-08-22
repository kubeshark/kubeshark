package tap

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/bradleyfalzon/tlsx"
	"github.com/up9inc/mizu/tap/api"
)

const checkTLSPacketAmount = 100

type tcpReaderDataMsg struct {
	bytes     []byte
	timestamp time.Time
}

type tcpID struct {
	srcIP   string
	dstIP   string
	srcPort string
	dstPort string
}

type ConnectionInfo struct {
	ClientIP   string
	ClientPort string
	ServerIP   string
	ServerPort string
	IsOutgoing bool
}

func (tid *tcpID) String() string {
	return fmt.Sprintf("%s->%s %s->%s", tid.srcIP, tid.dstIP, tid.srcPort, tid.dstPort)
}

/* tcpReader gets reads from a channel of bytes of tcp payload, and parses it into requests and responses.
 * The payload is written to the channel by a tcpStream object that is dedicated to one tcp connection.
 * An tcpReader object is unidirectional: it parses either a client stream or a server stream.
 * Implements io.Reader interface (Read)
 */
type tcpReader struct {
	ident              string
	tcpID              *api.TcpID
	isClient           bool
	isOutgoing         bool
	msgQueue           chan tcpReaderDataMsg // Channel of captured reassembled tcp payload
	data               []byte
	captureTime        time.Time
	parent             *tcpStream
	messageCount       uint
	packetsSeen        uint
	outboundLinkWriter *OutboundLinkWriter
	Emitter            api.Emitter
}

func (h *tcpReader) Read(p []byte) (int, error) {
	var msg tcpReaderDataMsg

	ok := true
	for ok && len(h.data) == 0 {
		msg, ok = <-h.msgQueue
		h.data = msg.bytes

		h.captureTime = msg.timestamp
		if len(h.data) > 0 {
			h.packetsSeen += 1
		}
		if h.packetsSeen < checkTLSPacketAmount && len(msg.bytes) > 5 { // packets with less than 5 bytes cause tlsx to panic
			clientHello := tlsx.ClientHello{}
			err := clientHello.Unmarshall(msg.bytes)
			if err == nil {
				fmt.Printf("Detected TLS client hello with SNI %s\n", clientHello.SNI)
				numericPort, _ := strconv.Atoi(h.tcpID.DstPort)
				h.outboundLinkWriter.WriteOutboundLink(h.tcpID.SrcIP, h.tcpID.DstIP, numericPort, clientHello.SNI, TLSProtocol)
			}
		}
	}
	if !ok || len(h.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, h.data)
	h.data = h.data[l:]
	return l, nil
}

func containsPort(ports []string, port string) bool {
	for _, x := range ports {
		if x == port {
			return true
		}
	}
	return false
}

func (h *tcpReader) run(wg *sync.WaitGroup) {
	defer wg.Done()
	var port string
	if h.isClient {
		port = h.tcpID.DstPort
	} else {
		port = h.tcpID.SrcPort
	}
	b := bufio.NewReader(h)
	for _, extension := range extensions {
		if containsPort(extension.Protocol.Ports, port) {
			extension.Dissector.Dissect(b, h.isClient, h.tcpID, h.Emitter)
		}
	}
}
