package tap

type OutboundLink struct {
	Src     string
	DstIP   string
	DstPort int
}

func NewOutboundLinkWriter() *OutboundLinkWriter {
	return &OutboundLinkWriter{
		OutChan: make(chan *OutboundLink),
	}
}

type OutboundLinkWriter struct {
	OutChan chan *OutboundLink
}

func (olw *OutboundLinkWriter) WriteOutboundLink(src string, DstIP string, DstPort int) {
	olw.OutChan <- &OutboundLink{
		Src: src,
		DstIP: DstIP,
		DstPort: DstPort,
	}
}

func (olw *OutboundLinkWriter) Stop() {
	close(olw.OutChan)
}
