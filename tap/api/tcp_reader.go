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

type TcpReader interface {
	Read(p []byte) (int, error)
	Close()
	Run(options *shared.TrafficFilteringOptions, wg *sync.WaitGroup)
	SendMsgIfNotClosed(msg TcpReaderDataMsg)
	GetReqResMatcher() RequestResponseMatcher
	GetIsClient() bool
	GetReadProgress() *ReadProgress
	GetParent() TcpStream
	GetTcpID() *TcpID
	GetCounterPair() *CounterPair
	GetCaptureTime() time.Time
	GetEmitter() Emitter
	GetIsClosed() bool
	GetExtension() *Extension
}

/* TcpReader gets reads from a channel of bytes of tcp payload, and parses it into requests and responses.
 * The payload is written to the channel by a tcpStream object that is dedicated to one tcp connection.
 * An TcpReader object is unidirectional: it parses either a client stream or a server stream.
 * Implements io.Reader interface (Read)
 */
type tcpReader struct {
	ident         string
	tcpID         *TcpID
	isClosed      bool
	isClient      bool
	isOutgoing    bool
	msgQueue      chan TcpReaderDataMsg // Channel of captured reassembled tcp payload
	data          []byte
	progress      *ReadProgress
	captureTime   time.Time
	parent        TcpStream
	packetsSeen   uint
	extension     *Extension
	emitter       Emitter
	counterPair   *CounterPair
	reqResMatcher RequestResponseMatcher
	sync.Mutex
}

func NewTcpReader(msgQueue chan TcpReaderDataMsg, progress *ReadProgress, ident string, tcpId *TcpID, captureTime time.Time, parent TcpStream, isClient bool, isOutgoing bool, extension *Extension, emitter Emitter, counterPair *CounterPair, reqResMatcher RequestResponseMatcher) TcpReader {
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

func (reader *tcpReader) Read(p []byte) (int, error) {
	var msg TcpReaderDataMsg

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

func (reader *tcpReader) Close() {
	reader.Lock()
	if !reader.isClosed {
		reader.isClosed = true
		close(reader.msgQueue)
	}
	reader.Unlock()
}

func (reader *tcpReader) Run(options *shared.TrafficFilteringOptions, wg *sync.WaitGroup) {
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

func (reader *tcpReader) SendMsgIfNotClosed(msg TcpReaderDataMsg) {
	reader.Lock()
	if !reader.isClosed {
		reader.msgQueue <- msg
	}
	reader.Unlock()
}

func (reader *tcpReader) GetReqResMatcher() RequestResponseMatcher {
	return reader.reqResMatcher
}

func (reader *tcpReader) GetIsClient() bool {
	return reader.isClient
}

func (reader *tcpReader) GetReadProgress() *ReadProgress {
	return reader.progress
}

func (reader *tcpReader) GetParent() TcpStream {
	return reader.parent
}

func (reader *tcpReader) GetTcpID() *TcpID {
	return reader.tcpID
}

func (reader *tcpReader) GetCounterPair() *CounterPair {
	return reader.counterPair
}

func (reader *tcpReader) GetCaptureTime() time.Time {
	return reader.captureTime
}

func (reader *tcpReader) GetEmitter() Emitter {
	return reader.emitter
}

func (reader *tcpReader) GetIsClosed() bool {
	return reader.isClosed
}

func (reader *tcpReader) GetExtension() *Extension {
	return reader.extension
}
