package tap

import (
	"os"
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
	id             int64
	isClosed       bool
	protocol       *api.Protocol
	isTapTarget    bool
	client         *tcpReader
	server         *tcpReader
	origin         api.Capture
	counterPairs   []*api.CounterPair
	reqResMatchers []api.RequestResponseMatcher
	createdAt      time.Time
	streamsMap     api.TcpStreamMap
	sync.Mutex
}

func NewTcpStream(isTapTarget bool, streamsMap api.TcpStreamMap, capture api.Capture) *tcpStream {
	return &tcpStream{
		isTapTarget: isTapTarget,
		streamsMap:  streamsMap,
		origin:      capture,
		createdAt:   time.Now(),
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

	t.client.close()
	t.server.close()
}

func (t *tcpStream) addCounterPair(counterPair *api.CounterPair) {
	t.counterPairs = append(t.counterPairs, counterPair)
}

func (t *tcpStream) addReqResMatcher(reqResMatcher api.RequestResponseMatcher) {
	t.reqResMatchers = append(t.reqResMatchers, reqResMatcher)
}

func (t *tcpStream) SetProtocol(protocol *api.Protocol) {
	t.protocol = protocol

	// Clean the buffers
	t.client.msgBufferMaster = make([]api.TcpReaderDataMsg, 0)
	t.server.msgBufferMaster = make([]api.TcpReaderDataMsg, 0)
}

func (t *tcpStream) GetOrigin() api.Capture {
	return t.origin
}

func (t *tcpStream) GetProtocol() *api.Protocol {
	return t.protocol
}

func (t *tcpStream) GetReqResMatchers() []api.RequestResponseMatcher {
	return t.reqResMatchers
}

func (t *tcpStream) GetIsTapTarget() bool {
	if os.Getenv("MIZU_TAPPER_DISABLE_TAP_TARGET") == "true" {
		return false
	}
	return t.isTapTarget
}

func (t *tcpStream) GetIsClosed() bool {
	return t.isClosed
}
