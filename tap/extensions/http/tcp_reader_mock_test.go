package http

import (
	"sync"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

type tcpReader struct {
	ident         string
	tcpID         *api.TcpID
	isClosed      bool
	isClient      bool
	isOutgoing    bool
	progress      *api.ReadProgress
	captureTime   time.Time
	parent        api.TcpStream
	extension     *api.Extension
	emitter       api.Emitter
	counterPair   *api.CounterPair
	reqResMatcher api.RequestResponseMatcher
	sync.Mutex
}

func NewTcpReader(progress *api.ReadProgress, ident string, tcpId *api.TcpID, captureTime time.Time, parent api.TcpStream, isClient bool, isOutgoing bool, extension *api.Extension, emitter api.Emitter, counterPair *api.CounterPair, reqResMatcher api.RequestResponseMatcher) api.TcpReader {
	return &tcpReader{
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
	return 0, nil
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
