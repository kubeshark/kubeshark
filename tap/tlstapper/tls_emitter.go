package tlstapper

import "github.com/kubeshark/kubeshark/tap/api"

type tlsEmitter struct {
	delegate  api.Emitter
	namespace string
}

func (e *tlsEmitter) Emit(item *api.OutputChannelItem) {
	item.Namespace = e.namespace
	e.delegate.Emit(item)
}
