package tlstapper

import (
	"io"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

type golangReader struct {
	key           string
	msgQueue      chan []byte
	data          []byte
	progress      *api.ReadProgress
	tcpID         *api.TcpID
	isClient      bool
	captureTime   time.Time
	extension     *api.Extension
	emitter       api.Emitter
	counterPair   *api.CounterPair
	parent        *tlsStream
	reqResMatcher api.RequestResponseMatcher
}

func NewGolangReader(extension *api.Extension, isClient bool, emitter api.Emitter, counterPair *api.CounterPair, stream *tlsStream, reqResMatcher api.RequestResponseMatcher) *golangReader {
	return &golangReader{
		msgQueue:      make(chan []byte, 1),
		progress:      &api.ReadProgress{},
		tcpID:         &api.TcpID{},
		isClient:      isClient,
		captureTime:   time.Now(),
		extension:     extension,
		emitter:       emitter,
		counterPair:   counterPair,
		parent:        stream,
		reqResMatcher: reqResMatcher,
	}
}

func (r *golangReader) send(b []byte) {
	r.captureTime = time.Now()
	r.msgQueue <- b
}

func (r *golangReader) Read(p []byte) (int, error) {
	var b []byte

	for len(r.data) == 0 {
		var ok bool
		select {
		case b, ok = <-r.msgQueue:
			if !ok {
				return 0, io.EOF
			}

			r.data = b
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

func (r *golangReader) GetReqResMatcher() api.RequestResponseMatcher {
	return r.reqResMatcher
}

func (r *golangReader) GetIsClient() bool {
	return r.isClient
}

func (r *golangReader) GetReadProgress() *api.ReadProgress {
	return r.progress
}

func (r *golangReader) GetParent() api.TcpStream {
	return r.parent
}

func (r *golangReader) GetTcpID() *api.TcpID {
	return r.tcpID
}

func (r *golangReader) GetCounterPair() *api.CounterPair {
	return r.counterPair
}

func (r *golangReader) GetCaptureTime() time.Time {
	return r.captureTime
}

func (r *golangReader) GetEmitter() api.Emitter {
	return r.emitter
}

func (r *golangReader) GetIsClosed() bool {
	return false
}
