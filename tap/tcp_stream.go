package tap

import (
	"sync"
	"time"

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
	clients         []*tcpReader
	servers         []*tcpReader
	origin          api.Capture
	reqResMatchers  []api.RequestResponseMatcher
	createdAt       time.Time
	streamsMap      api.TcpStreamMap
	sync.Mutex
}

func NewTcpStream(isTapTarget bool, streamsMap api.TcpStreamMap, capture api.Capture) *tcpStream {
	return &tcpStream{
		isTapTarget:     isTapTarget,
		protoIdentifier: &api.ProtoIdentifier{},
		streamsMap:      streamsMap,
		origin:          capture,
	}
}

func (t *tcpStream) getId() int64 {
	return t.id
}

func (t *tcpStream) setId(id int64) {
	t.id = id
}

func (t *tcpStream) close() {
	t.Lock()
	defer t.Unlock()

	if t.isClosed {
		return
	}

	t.isClosed = true

	t.streamsMap.Delete(t.id)

	for i := range t.clients {
		reader := t.clients[i]
		reader.close()
	}
	for i := range t.servers {
		reader := t.servers[i]
		reader.close()
	}
}

func (t *tcpStream) addClient(reader *tcpReader) {
	t.clients = append(t.clients, reader)
}

func (t *tcpStream) addServer(reader *tcpReader) {
	t.servers = append(t.servers, reader)
}

func (t *tcpStream) getClients() []*tcpReader {
	return t.clients
}

func (t *tcpStream) getServers() []*tcpReader {
	return t.servers
}

func (t *tcpStream) getClient(index int) *tcpReader {
	return t.clients[index]
}

func (t *tcpStream) getServer(index int) *tcpReader {
	return t.servers[index]
}

func (t *tcpStream) addReqResMatcher(reqResMatcher api.RequestResponseMatcher) {
	t.reqResMatchers = append(t.reqResMatchers, reqResMatcher)
}

func (t *tcpStream) SetProtocol(protocol *api.Protocol) {
	t.Lock()
	defer t.Unlock()

	if t.protoIdentifier.IsClosedOthers {
		return
	}

	t.protoIdentifier.Protocol = protocol

	for i := range t.clients {
		reader := t.clients[i]
		if reader.GetExtension().Protocol != t.protoIdentifier.Protocol {
			reader.close()
		}
	}
	for i := range t.servers {
		reader := t.servers[i]
		if reader.GetExtension().Protocol != t.protoIdentifier.Protocol {
			reader.close()
		}
	}

	t.protoIdentifier.IsClosedOthers = true
}

func (t *tcpStream) GetOrigin() api.Capture {
	return t.origin
}

func (t *tcpStream) GetProtoIdentifier() *api.ProtoIdentifier {
	return t.protoIdentifier
}

func (t *tcpStream) GetReqResMatchers() []api.RequestResponseMatcher {
	return t.reqResMatchers
}

func (t *tcpStream) GetIsTapTarget() bool {
	return t.isTapTarget
}

func (t *tcpStream) GetIsClosed() bool {
	return t.isClosed
}
