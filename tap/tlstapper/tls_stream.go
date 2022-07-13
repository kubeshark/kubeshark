package tlstapper

import "github.com/up9inc/mizu/tap/api"

type tlsStream struct {
	reader   *tlsReader
	protocol *api.Protocol
}

func (t *tlsStream) GetOrigin() api.Capture {
	return api.Ebpf
}

func (t *tlsStream) SetProtocol(protocol *api.Protocol) {
	t.protocol = protocol
}

func (t *tlsStream) GetReqResMatchers() []api.RequestResponseMatcher {
	return []api.RequestResponseMatcher{t.reader.reqResMatcher}
}

func (t *tlsStream) GetIsTapTarget() bool {
	return true
}

func (t *tlsStream) GetIsClosed() bool {
	return false
}
