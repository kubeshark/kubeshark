package tlstapper

import (
	"io"
	"sync"
	"time"

	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap/api"
)

type tlsReader struct {
	key           string
	chunks        chan api.TlsChunk
	data          []byte
	doneHandler   func(r *tlsReader)
	progress      *api.ReadProgress
	tcpID         *api.TcpID
	isClosed      bool
	isClient      bool
	captureTime   time.Time
	parent        api.TcpStream
	extension     *api.Extension
	emitter       api.Emitter
	counterPair   *api.CounterPair
	reqResMatcher api.RequestResponseMatcher
}

func NewTlsReader(key string, doneHandler func(r *tlsReader), isClient bool, stream api.TcpStream) api.TcpReader {
	return &tlsReader{
		key:         key,
		chunks:      make(chan api.TlsChunk, 1),
		doneHandler: doneHandler,
		parent:      stream,
	}
}

func (r *tlsReader) Read(p []byte) (int, error) {
	var chunk api.TlsChunk

	for len(r.data) == 0 {
		var ok bool
		select {
		case chunk, ok = <-r.chunks:
			if !ok {
				return 0, io.EOF
			}

			r.data = chunk.GetRecordedData()
		case <-time.After(time.Second * 3):
			r.doneHandler(r)
			return 0, io.EOF
		}

		if len(r.data) > 0 {
			break
		}
	}

	l := copy(p, r.data)
	r.data = r.data[l:]
	r.progress.Feed(l)

	return l, nil
}

func (r *tlsReader) Close() {
	r.doneHandler(r)
}

func (r *tlsReader) Run(options *shared.TrafficFilteringOptions, wg *sync.WaitGroup) {}

func (r *tlsReader) SendMsgIfNotClosed(msg api.TcpReaderDataMsg) {}

func (r *tlsReader) SendChunk(chunk api.TlsChunk) {
	r.chunks <- chunk
}

func (r *tlsReader) GetReqResMatcher() api.RequestResponseMatcher {
	return r.reqResMatcher
}

func (r *tlsReader) GetIsClient() bool {
	return r.isClient
}

func (r *tlsReader) GetReadProgress() *api.ReadProgress {
	return r.progress
}

func (r *tlsReader) GetParent() api.TcpStream {
	return r.parent
}

func (r *tlsReader) GetTcpID() *api.TcpID {
	return r.tcpID
}

func (r *tlsReader) GetCounterPair() *api.CounterPair {
	return r.counterPair
}

func (r *tlsReader) GetCaptureTime() time.Time {
	return r.captureTime
}

func (r *tlsReader) GetEmitter() api.Emitter {
	return r.emitter
}

func (r *tlsReader) GetIsClosed() bool {
	return r.isClosed
}

func (r *tlsReader) GetExtension() *api.Extension {
	return r.extension
}

func (r *tlsReader) SetTcpID(tcpID *api.TcpID) {
	r.tcpID = tcpID
}

func (r *tlsReader) SetCaptureTime(captureTime time.Time) {
	r.captureTime = captureTime
}

func (r *tlsReader) SetEmitter(emitter api.Emitter) {
	r.emitter = emitter
}
