package tap

import (
	"github.com/google/gopacket" // pulls in all layers decoders
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

/* It's a connection (bidirectional)
 * Implements gopacket.reassembly.Stream interface (Accept, ReassembledSG, ReassemblyComplete)
 * ReassembledSG gets called when new reassembled data is ready (i.e. bytes in order, no duplicates, complete)
 * In our implementation, we pass information from ReassembledSG to the httpReader through a shared channel.
 */
type tcpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}
