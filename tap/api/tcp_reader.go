package api

import (
	"bufio"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/up9inc/mizu/shared/logger"
)

type TcpReaderDataMsg struct {
	bytes     []byte
	timestamp time.Time
}

/* TcpReader gets reads from a channel of bytes of tcp payload, and parses it into requests and responses.
 * The payload is written to the channel by a tcpStream object that is dedicated to one tcp connection.
 * An TcpReader object is unidirectional: it parses either a client stream or a server stream.
 * Implements io.Reader interface (Read)
 */
type TcpReader struct {
	Ident         string
	TcpID         *TcpID
	isClosed      bool
	IsClient      bool
	IsOutgoing    bool
	MsgQueue      chan TcpReaderDataMsg // Channel of captured reassembled tcp payload
	data          []byte
	Progress      *ReadProgress
	SuperTimer    *SuperTimer
	Parent        *TcpStream
	packetsSeen   uint
	Extension     *Extension
	Emitter       Emitter
	CounterPair   *CounterPair
	ReqResMatcher RequestResponseMatcher
	sync.Mutex
}

func (h *TcpReader) Read(p []byte) (int, error) {
	var msg TcpReaderDataMsg

	ok := true
	for ok && len(h.data) == 0 {
		msg, ok = <-h.MsgQueue
		h.data = msg.bytes

		h.SuperTimer.CaptureTime = msg.timestamp
		if len(h.data) > 0 {
			h.packetsSeen += 1
		}
	}
	if !ok || len(h.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, h.data)
	h.data = h.data[l:]
	h.Progress.Feed(l)

	return l, nil
}

func (h *TcpReader) Close() {
	h.Lock()
	if !h.isClosed {
		h.isClosed = true
		close(h.MsgQueue)
	}
	h.Unlock()
}

func (h *TcpReader) Run(filteringOptions *TrafficFilteringOptions, wg *sync.WaitGroup) {
	defer wg.Done()
	b := bufio.NewReader(h)
	err := h.Extension.Dissector.Dissect(b, h.Progress, h.Parent.Origin, h.IsClient, h.TcpID, h.CounterPair, h.SuperTimer, h.Parent.SuperIdentifier, h.Emitter, filteringOptions, h.ReqResMatcher)
	if err != nil {
		_, err = io.Copy(ioutil.Discard, b)
		if err != nil {
			logger.Log.Errorf("%v", err)
		}
	}
}
