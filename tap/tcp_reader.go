package tap

import (
	"bufio"
	"bytes"
	"io"
	"sync"
	"time"

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
	msgBuffer     *bytes.Buffer
	data          []byte
	progress      *api.ReadProgress
	captureTime   time.Time
	parent        *tcpStream
	packetsSeen   uint
	emitter       api.Emitter
	counterPair   *api.CounterPair
	reqResMatcher api.RequestResponseMatcher
	sync.Mutex
}

func NewTcpReader(msgQueue chan api.TcpReaderDataMsg, progress *api.ReadProgress, ident string, tcpId *api.TcpID, captureTime time.Time, parent *tcpStream, isClient bool, isOutgoing bool, emitter api.Emitter) *tcpReader {
	return &tcpReader{
		msgQueue:    msgQueue,
		msgBuffer:   bytes.NewBuffer(make([]byte, 0)),
		progress:    progress,
		ident:       ident,
		tcpID:       tcpId,
		captureTime: captureTime,
		parent:      parent,
		isClient:    isClient,
		isOutgoing:  isOutgoing,
		emitter:     emitter,
	}
}

func (reader *tcpReader) run(options *api.TrafficFilteringOptions, wg *sync.WaitGroup) {
	defer wg.Done()
	for i, extension := range extensions {
		reader.reqResMatcher = reader.parent.reqResMatchers[i]
		reader.counterPair = reader.parent.counterPairs[i]
		b := bufio.NewReader(reader)
		extension.Dissector.Dissect(b, reader, options) //nolint
		if reader.isProtocolIdentified() {
			break
		}
		reader.rewind()
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

func (reader *tcpReader) isProtocolIdentified() bool {
	return reader.parent.protocol != nil
}

func (reader *tcpReader) rewind() {
	// Reset the data and msgBuffer from the master record
	buffer := reader.msgBuffer.Bytes()
	reader.data = make([]byte, len(buffer))
	copy(reader.data, buffer)

	// Reset the read progress
	reader.progress.Reset()
}

func (reader *tcpReader) Read(p []byte) (int, error) {
	var msg api.TcpReaderDataMsg
	ok := true
	for ok && len(reader.data) == 0 {
		msg, ok = <-reader.msgQueue
		if msg != nil {
			reader.data = msg.GetBytes()
			reader.captureTime = msg.GetTimestamp()

			if !reader.isProtocolIdentified() {
				reader.msgBuffer.Write(reader.data)
			}
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
