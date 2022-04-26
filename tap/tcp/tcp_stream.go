package tcp

import (
	"sync"
	"time"

	"github.com/google/gopacket/layers" // pulls in all layers decoders
	"github.com/up9inc/mizu/tap/api"
)

/* It's a connection (bidirectional)
 * Implements gopacket.reassembly.Stream interface (Accept, ReassembledSG, ReassemblyComplete)
 * ReassembledSG gets called when new reassembled data is ready (i.e. bytes in order, no duplicates, complete)
 * In our implementation, we pass information from ReassembledSG to the TcpReader through a shared channel.
 */
type tcpStream struct {
	id              int64
	isClosed        bool
	protoIdentifier *api.ProtoIdentifier
	isTapTarget     bool
	clients         []api.TcpReader
	servers         []api.TcpReader
	origin          api.Capture
	reqResMatcher   api.RequestResponseMatcher
	createdAt       time.Time
	streamsMap      api.TcpStreamMap
	sync.Mutex
}

func NewTcpStream(tcp *layers.TCP, isTapTarget bool, streamsMap api.TcpStreamMap, capture api.Capture) api.TcpStream {
	return &tcpStream{
		isTapTarget:     isTapTarget,
		protoIdentifier: &api.ProtoIdentifier{},
		streamsMap:      streamsMap,
		origin:          capture,
	}
}

func NewTcpStreamDummy(capture api.Capture) api.TcpStream {
	return &tcpStream{
		origin:          capture,
		protoIdentifier: &api.ProtoIdentifier{},
	}
}

func (t *tcpStream) Close() {
	t.Lock()
	defer t.Unlock()

	if t.isClosed {
		return
	}

	t.isClosed = true

	t.streamsMap.Delete(t.id)

	for i := range t.clients {
		reader := t.clients[i]
		reader.Close()
	}
	for i := range t.servers {
		reader := t.servers[i]
		reader.Close()
	}
}

func (t *tcpStream) CloseOtherProtocolDissectors(protocol *api.Protocol) {
	t.Lock()
	defer t.Unlock()

	if t.protoIdentifier.IsClosedOthers {
		return
	}

	t.protoIdentifier.Protocol = protocol

	for i := range t.clients {
		reader := t.clients[i]
		if reader.GetExtension().Protocol != t.protoIdentifier.Protocol {
			reader.Close()
		}
	}
	for i := range t.servers {
		reader := t.servers[i]
		if reader.GetExtension().Protocol != t.protoIdentifier.Protocol {
			reader.Close()
		}
	}

	t.protoIdentifier.IsClosedOthers = true
}

func (t *tcpStream) AddClient(reader api.TcpReader) {
	t.clients = append(t.clients, reader)
}

func (t *tcpStream) AddServer(reader api.TcpReader) {
	t.servers = append(t.servers, reader)
}

func (t *tcpStream) GetClients() []api.TcpReader {
	return t.clients
}

func (t *tcpStream) GetServers() []api.TcpReader {
	return t.servers
}

func (t *tcpStream) GetClient(index int) api.TcpReader {
	return t.clients[index]
}

func (t *tcpStream) GetServer(index int) api.TcpReader {
	return t.servers[index]
}

func (t *tcpStream) GetOrigin() api.Capture {
	return t.origin
}

func (t *tcpStream) GetProtoIdentifier() *api.ProtoIdentifier {
	return t.protoIdentifier
}

func (t *tcpStream) GetReqResMatcher() api.RequestResponseMatcher {
	return t.reqResMatcher
}

func (t *tcpStream) GetIsTapTarget() bool {
	return t.isTapTarget
}

func (t *tcpStream) GetIsClosed() bool {
	return t.isClosed
}

func (t *tcpStream) GetId() int64 {
	return t.id
}

func (t *tcpStream) SetId(id int64) {
	t.id = id
}
