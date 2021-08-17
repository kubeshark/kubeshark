package tap

import (
	"time"

	"github.com/google/gopacket" // pulls in all layers decoders
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

type tcpReaderDataMsg struct {
	bytes     []byte
	timestamp time.Time
}

type tcpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}
