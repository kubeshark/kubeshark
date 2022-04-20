package api

import (
	"bufio"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/up9inc/mizu/shared"
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

func (reader *TcpReader) Read(p []byte) (int, error) {
	var msg TcpReaderDataMsg

	ok := true
	for ok && len(reader.data) == 0 {
		msg, ok = <-reader.MsgQueue
		reader.data = msg.bytes

		reader.SuperTimer.CaptureTime = msg.timestamp
		if len(reader.data) > 0 {
			reader.packetsSeen += 1
		}
	}
	if !ok || len(reader.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, reader.data)
	reader.data = reader.data[l:]
	reader.Progress.Feed(l)

	return l, nil
}

func (reader *TcpReader) Close() {
	reader.Lock()
	if !reader.isClosed {
		reader.isClosed = true
		close(reader.MsgQueue)
	}
	reader.Unlock()
}

func (reader *TcpReader) Run(options *shared.TrafficFilteringOptions, wg *sync.WaitGroup) {
	defer wg.Done()
	b := bufio.NewReader(reader)
	err := reader.Extension.Dissector.Dissect(b, reader, options)
	if err != nil {
		_, err = io.Copy(ioutil.Discard, reader)
		if err != nil {
			logger.Log.Errorf("%v", err)
		}
	}
}
