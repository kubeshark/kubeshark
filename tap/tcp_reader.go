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

/* TcpReader gets reads from a channel of bytes of tcp payload, and parses it into requests and responses.
 * The payload is written to the channel by a tcpStream object that is dedicated to one tcp connection.
 * An TcpReader object is unidirectional: it parses either a client stream or a server stream.
 * Implements io.Reader interface (Read)
 */
type tcpReader struct {
	ident         string
	tcpID         *api.TcpID
	isClosed      bool
	isClient      bool
	isOutgoing    bool
	msgQueue      chan api.TcpReaderDataMsg // Channel of captured reassembled tcp payload
	data          []byte
	progress      *api.ReadProgress
	captureTime   time.Time
	parent        api.TcpStream
	packetsSeen   uint
	extension     *api.Extension
	emitter       api.Emitter
	counterPair   *api.CounterPair
	reqResMatcher api.RequestResponseMatcher
	sync.Mutex
}

func NewTcpReader(msgQueue chan api.TcpReaderDataMsg, progress *api.ReadProgress, ident string, tcpId *api.TcpID, captureTime time.Time, parent api.TcpStream, isClient bool, isOutgoing bool, extension *api.Extension, emitter api.Emitter, counterPair *api.CounterPair, reqResMatcher api.RequestResponseMatcher) api.TcpReader {
	return &tcpReader{
		msgQueue:      msgQueue,
		progress:      progress,
		ident:         ident,
		tcpID:         tcpId,
		captureTime:   captureTime,
		parent:        parent,
		isClient:      isClient,
		isOutgoing:    isOutgoing,
		extension:     extension,
		emitter:       emitter,
		counterPair:   counterPair,
		reqResMatcher: reqResMatcher,
	}
}

func (reader *tcpReader) run(options *api.TrafficFilteringOptions, wg *sync.WaitGroup) {
	defer wg.Done()
	b := bufio.NewReader(reader)
	err := reader.extension.Dissector.Dissect(b, reader, options)
	if err != nil {
		_, err = io.Copy(ioutil.Discard, reader)
		if err != nil {
			logger.Log.Errorf("%v", err)
		}
	}
}

func (reader *tcpReader) close() {
	reader.Lock()
	if !reader.isClosed {
		reader.isClosed = true
		close(reader.msgQueue)
	}
	reader.Unlock()
}

func (reader *tcpReader) sendMsgIfNotClosed(msg api.TcpReaderDataMsg) {
	reader.Lock()
	if !reader.isClosed {
		reader.msgQueue <- msg
	}
	reader.Unlock()
}

func (reader *tcpReader) Read(p []byte) (int, error) {
	var msg api.TcpReaderDataMsg

	ok := true
	for ok && len(reader.data) == 0 {
		msg, ok = <-reader.msgQueue
		if msg != nil {
			reader.data = msg.GetBytes()
			reader.captureTime = msg.GetTimestamp()
		}

		if len(reader.data) > 0 {
			reader.packetsSeen += 1
		}
	}
	if !ok || len(reader.data) == 0 {
		return 0, io.EOF
	}

	l := copy(p, reader.data)
	reader.data = reader.data[l:]
	reader.progress.Feed(l)

	return l, nil
}

func (reader *tcpReader) GetReqResMatcher() api.RequestResponseMatcher {
	return reader.reqResMatcher
}

func (reader *tcpReader) GetIsClient() bool {
	return reader.isClient
}

func (reader *tcpReader) GetReadProgress() *api.ReadProgress {
	return reader.progress
}

func (reader *tcpReader) GetParent() api.TcpStream {
	return reader.parent
}

func (reader *tcpReader) GetTcpID() *api.TcpID {
	return reader.tcpID
}

func (reader *tcpReader) GetCounterPair() *api.CounterPair {
	return reader.counterPair
}

func (reader *tcpReader) GetCaptureTime() time.Time {
	return reader.captureTime
}

func (reader *tcpReader) GetEmitter() api.Emitter {
	return reader.emitter
}

func (reader *tcpReader) GetIsClosed() bool {
	return reader.isClosed
}

func (reader *tcpReader) GetExtension() *api.Extension {
	return reader.extension
}
