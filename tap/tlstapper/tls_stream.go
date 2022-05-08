package tlstapper

import "github.com/up9inc/mizu/tap/api"

type tlsStream struct {
	reader          *tlsReader
	protoIdentifier *api.ProtoIdentifier
}

func (t *tlsStream) GetOrigin() api.Capture {
	return api.Ebpf
}

func (t *tlsStream) GetProtoIdentifier() *api.ProtoIdentifier {
	return t.protoIdentifier
}

func (t *tlsStream) SetProtocol(protocol *api.Protocol) {
	t.protoIdentifier.Protocol = protocol
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
