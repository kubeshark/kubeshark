package tap

import (
	"bufio"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
)

type tcpReaderDataMsg struct {
	bytes     []byte
	timestamp time.Time
}

type ConnectionInfo struct {
	ClientIP   string
	ClientPort string
	ServerIP   string
	ServerPort string
	IsOutgoing bool
}

/* tcpReader gets reads from a channel of bytes of tcp payload, and parses it into requests and responses.
 * The payload is written to the channel by a tcpStream object that is dedicated to one tcp connection.
 * An tcpReader object is unidirectional: it parses either a client stream or a server stream.
 * Implements io.Reader interface (Read)
 */
type tcpReader struct {
	ident         string
	tcpID         *api.TcpID
	isClosed      bool
	isClient      bool
	isOutgoing    bool
	msgQueue      chan tcpReaderDataMsg // Channel of captured reassembled tcp payload
	data          []byte
	progress      *api.ReadProgress
	superTimer    *api.SuperTimer
	parent        *tcpStream
	packetsSeen   uint
	extension     *api.Extension
	emitter       api.Emitter
	counterPair   *api.CounterPair
	reqResMatcher api.RequestResponseMatcher
	sync.Mutex
}

func (h *tcpReader) Read(p []byte) (int, error) {
	var msg tcpReaderDataMsg

	ok := true
	for ok && len(h.data) == 0 {
		msg, ok = <-h.msgQueue
		h.data = msg.bytes

		h.superTimer.CaptureTime = msg.timestamp
		if len(h.data) > 0 {
			h.packetsSeen += 1
		}
	}
	if !ok || len(h.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, h.data)
	h.data = h.data[l:]
	h.progress.Feed(l)

	return l, nil
}

func (h *tcpReader) Close() {
	h.Lock()
	if !h.isClosed {
		h.isClosed = true
		close(h.msgQueue)
	}
	h.Unlock()
}

func (h *tcpReader) run(wg *sync.WaitGroup) {
	defer wg.Done()
	b := bufio.NewReader(h)
	err := h.extension.Dissector.Dissect(b, h.progress, h.parent.origin, h.isClient, h.tcpID, h.counterPair, h.superTimer, h.parent.superIdentifier, h.emitter, filteringOptions, h.reqResMatcher)
	if err != nil {
		_, err = io.Copy(ioutil.Discard, b)
		if err != nil {
			logger.Log.Errorf("%v", err)
		}
	}
}
