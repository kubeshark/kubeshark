package tap

type OutboundLinkProtocol string

const (
	TLSProtocol OutboundLinkProtocol = "tls"
)

type OutboundLink struct {
	Src     string
	DstIP   string
	DstPort int
	SuggestedResolvedName string
	SuggestedProtocol OutboundLinkProtocol
}

func NewOutboundLinkWriter() *OutboundLinkWriter {
	return &OutboundLinkWriter{
		OutChan: make(chan *OutboundLink),
	}
}

type OutboundLinkWriter struct {
	OutChan chan *OutboundLink
}

func (olw *OutboundLinkWriter) WriteOutboundLink(src string, DstIP string, DstPort int, SuggestedResolvedName string, SuggestedProtocol OutboundLinkProtocol) {
	olw.OutChan <- &OutboundLink{
		Src: src,
		DstIP: DstIP,
		DstPort: DstPort,
		SuggestedResolvedName: SuggestedResolvedName,
		SuggestedProtocol: SuggestedProtocol,
	}
}

func (olw *OutboundLinkWriter) Stop() {
	close(olw.OutChan)
}
